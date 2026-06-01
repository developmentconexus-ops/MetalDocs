package gotenberg

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	maxErrorBodyBytes = 4 * 1024
	maxPDFBodyBytes   = 64 * 1024 * 1024
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) (*Client, error) {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		return nil, errors.New("gotenberg: baseURL required")
	}
	parsed, err := url.ParseRequestURI(baseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, errors.New("gotenberg: baseURL must be an absolute http(s) URL")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, errors.New("gotenberg: baseURL must use http or https")
	}
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}, nil
}

// ConvertHTMLToPDF sends an HTML document plus an auxiliary stylesheet to
// Gotenberg's Chromium route and returns the rendered PDF bytes.
func (c *Client) ConvertHTMLToPDF(ctx context.Context, htmlBytes []byte, cssBytes []byte) ([]byte, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	htmlPart, err := writer.CreateFormFile("files", "index.html")
	if err != nil {
		return nil, fmt.Errorf("gotenberg: create html form file: %w", err)
	}
	if _, err := htmlPart.Write(htmlBytes); err != nil {
		return nil, fmt.Errorf("gotenberg: write html content: %w", err)
	}

	if len(cssBytes) > 0 {
		cssPart, err := writer.CreateFormFile("files", "style.css")
		if err != nil {
			return nil, fmt.Errorf("gotenberg: create css form file: %w", err)
		}
		if _, err := cssPart.Write(cssBytes); err != nil {
			return nil, fmt.Errorf("gotenberg: write css content: %w", err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("gotenberg: close multipart: %w", err)
	}

	url := c.baseURL + "/forms/chromium/convert/html"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &body)
	if err != nil {
		return nil, fmt.Errorf("gotenberg: create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gotenberg: html request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		payload, err := readLimitedBody(resp.Body, maxErrorBodyBytes)
		if err != nil {
			return nil, fmt.Errorf("gotenberg: read error response: %w", err)
		}
		return nil, fmt.Errorf("gotenberg: html conversion returned status %d: %s", resp.StatusCode, string(payload))
	}

	pdfBytes, err := readLimitedBody(resp.Body, maxPDFBodyBytes)
	if err != nil {
		return nil, fmt.Errorf("gotenberg: read pdf response: %w", err)
	}
	return pdfBytes, nil
}

func (c *Client) ConvertDocxToPDF(ctx context.Context, docxContent []byte) ([]byte, error) {
	return c.ConvertDocxToPDFWithOptions(ctx, docxContent, "", false)
}

// paperDimensionsInches maps a paper-size name to Gotenberg form dimensions.
// An unknown/empty name yields ok=false, letting LibreOffice use the docx's
// own page geometry.
func paperDimensionsInches(paperSize string) (width float64, height float64, ok bool) {
	switch paperSize {
	case "A4":
		return 8.27, 11.69, true
	case "Letter":
		return 8.5, 11.0, true
	default:
		return 0, 0, false
	}
}

// ConvertDocxToPDFWithOptions converts a docx to PDF via the LibreOffice route.
// When paperSize is "A4"/"Letter" it forces page dimensions (and landscape);
// otherwise the docx's own page setup is preserved.
func (c *Client) ConvertDocxToPDFWithOptions(ctx context.Context, docxContent []byte, paperSize string, landscape bool) ([]byte, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("files", "document.docx")
	if err != nil {
		return nil, fmt.Errorf("gotenberg: create form file: %w", err)
	}
	if _, err := part.Write(docxContent); err != nil {
		return nil, fmt.Errorf("gotenberg: write docx content: %w", err)
	}
	if w, h, ok := paperDimensionsInches(paperSize); ok {
		_ = writer.WriteField("paperWidth", strconv.FormatFloat(w, 'f', -1, 64))
		_ = writer.WriteField("paperHeight", strconv.FormatFloat(h, 'f', -1, 64))
		if landscape {
			_ = writer.WriteField("landscape", "true")
		}
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("gotenberg: close multipart: %w", err)
	}

	url := c.baseURL + "/forms/libreoffice/convert"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &body)
	if err != nil {
		return nil, fmt.Errorf("gotenberg: create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gotenberg: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, err := readLimitedBody(resp.Body, maxErrorBodyBytes)
		if err != nil {
			return nil, fmt.Errorf("gotenberg: read error response: %w", err)
		}
		return nil, fmt.Errorf("gotenberg: status %d: %s", resp.StatusCode, string(respBody))
	}

	pdfBytes, err := readLimitedBody(resp.Body, maxPDFBodyBytes)
	if err != nil {
		return nil, fmt.Errorf("gotenberg: read pdf response: %w", err)
	}
	return pdfBytes, nil
}

func readLimitedBody(body io.Reader, maxBytes int64) ([]byte, error) {
	limited := io.LimitReader(body, maxBytes+1)
	payload, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if int64(len(payload)) > maxBytes {
		return nil, fmt.Errorf("body exceeds %d-byte limit", maxBytes)
	}
	return payload, nil
}
