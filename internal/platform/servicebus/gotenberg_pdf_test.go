package servicebus

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"testing"
)

type fakeObjectStore struct {
	objects map[string][]byte
	saved   map[string][]byte
}

func newFakeObjectStore() *fakeObjectStore {
	return &fakeObjectStore{objects: map[string][]byte{}, saved: map[string][]byte{}}
}

func (f *fakeObjectStore) Open(_ context.Context, key string) (io.ReadCloser, error) {
	b, ok := f.objects[key]
	if !ok {
		return nil, errors.New("object not found: " + key)
	}
	return io.NopCloser(bytes.NewReader(b)), nil
}

func (f *fakeObjectStore) Save(_ context.Context, key string, content []byte) error {
	f.saved[key] = append([]byte(nil), content...)
	return nil
}

type fakeDocxConverter struct {
	pdf          []byte
	gotDocx      []byte
	gotPaperSize string
	gotLandscape bool
	err          error
}

func (f *fakeDocxConverter) ConvertDocxToPDFWithOptions(_ context.Context, docx []byte, paperSize string, landscape bool) ([]byte, error) {
	f.gotDocx = append([]byte(nil), docx...)
	f.gotPaperSize = paperSize
	f.gotLandscape = landscape
	if f.err != nil {
		return nil, f.err
	}
	return f.pdf, nil
}

func TestGotenbergPDFClient_ConvertPDF_HappyPath(t *testing.T) {
	docx := []byte("PK\x03\x04 fake docx bytes")
	pdf := []byte("%PDF-1.7 fake pdf bytes")
	store := newFakeObjectStore()
	store.objects["tenants/t1/revisions/r1/frozen.docx"] = docx
	conv := &fakeDocxConverter{pdf: pdf}

	client := NewGotenbergPDFClient(store, conv)

	res, err := client.ConvertPDF(context.Background(), ConvertPDFRequest{
		DocxKey:   "tenants/t1/revisions/r1/frozen.docx",
		OutputKey: "tenants/t1/revisions/r1/final.pdf",
	})
	if err != nil {
		t.Fatalf("ConvertPDF: unexpected error: %v", err)
	}

	if !bytes.Equal(conv.gotDocx, docx) {
		t.Errorf("converter received wrong docx bytes")
	}
	gotPDF, ok := store.saved["tenants/t1/revisions/r1/final.pdf"]
	if !ok {
		t.Fatalf("pdf was not saved to OutputKey")
	}
	if !bytes.Equal(gotPDF, pdf) {
		t.Errorf("saved pdf bytes mismatch")
	}
	if res.OutputKey != "tenants/t1/revisions/r1/final.pdf" {
		t.Errorf("OutputKey = %q", res.OutputKey)
	}
	wantHash := hex.EncodeToString(sha256Sum(pdf))
	if res.ContentHash != wantHash {
		t.Errorf("ContentHash = %q, want %q", res.ContentHash, wantHash)
	}
	if res.SizeBytes != int64(len(pdf)) {
		t.Errorf("SizeBytes = %d, want %d", res.SizeBytes, len(pdf))
	}
}

func TestGotenbergPDFClient_ConvertPDF_ThreadsRenderOpts(t *testing.T) {
	store := newFakeObjectStore()
	store.objects["in.docx"] = []byte("docx")
	conv := &fakeDocxConverter{pdf: []byte("%PDF")}
	client := NewGotenbergPDFClient(store, conv)

	_, err := client.ConvertPDF(context.Background(), ConvertPDFRequest{
		DocxKey:    "in.docx",
		OutputKey:  "out.pdf",
		RenderOpts: &PDFRenderOpts{PaperSize: PaperSizeLetter, Landscape: true},
	})
	if err != nil {
		t.Fatalf("ConvertPDF: unexpected error: %v", err)
	}
	if conv.gotPaperSize != "Letter" || !conv.gotLandscape {
		t.Errorf("render opts not threaded: paperSize=%q landscape=%v", conv.gotPaperSize, conv.gotLandscape)
	}
}

func TestGotenbergPDFClient_ConvertPDF_OpenError(t *testing.T) {
	store := newFakeObjectStore() // empty — DocxKey missing
	conv := &fakeDocxConverter{pdf: []byte("x")}
	client := NewGotenbergPDFClient(store, conv)

	_, err := client.ConvertPDF(context.Background(), ConvertPDFRequest{
		DocxKey:   "missing.docx",
		OutputKey: "out.pdf",
	})
	if err == nil {
		t.Fatalf("expected error when docx is missing")
	}
	if len(store.saved) != 0 {
		t.Errorf("nothing should be saved on open error")
	}
}

func TestGotenbergPDFClient_ConvertPDF_ConvertError(t *testing.T) {
	store := newFakeObjectStore()
	store.objects["in.docx"] = []byte("docx")
	conv := &fakeDocxConverter{err: errors.New("gotenberg down")}
	client := NewGotenbergPDFClient(store, conv)

	_, err := client.ConvertPDF(context.Background(), ConvertPDFRequest{
		DocxKey:   "in.docx",
		OutputKey: "out.pdf",
	})
	if err == nil {
		t.Fatalf("expected error when conversion fails")
	}
	if len(store.saved) != 0 {
		t.Errorf("nothing should be saved on convert error")
	}
}

func sha256Sum(b []byte) []byte {
	sum := sha256.Sum256(b)
	return sum[:]
}
