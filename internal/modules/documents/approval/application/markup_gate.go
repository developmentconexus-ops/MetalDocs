package application

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
)

// ErrUnresolvedTrackedChanges is returned by ScanForUnresolvedMarkup when the
// docx buffer contains an unaccepted/unrejected tracked-change element
// (w:ins, w:del, w:pPrChange). Markup must never persist into a frozen/signed
// artifact (spec.md §2.2 point 2).
var ErrUnresolvedTrackedChanges = errors.New("approval: unresolved tracked changes present in document")

// ErrUnstrippedComments is returned by ScanForUnresolvedMarkup when the docx
// buffer still contains comment-anchor marks (w:commentRangeStart,
// w:commentReference) that were not stripped before freeze.
var ErrUnstrippedComments = errors.New("approval: unstripped comment marks present in document")

// ErrInvalidDocxBuffer is returned when the supplied bytes are not a readable
// zip archive, or the archive has no word/document.xml entry. Fail-closed:
// an unreadable buffer is never treated as clean.
var ErrInvalidDocxBuffer = errors.New("approval: invalid or unreadable docx buffer")

// documentXMLPath is the fixed OOXML part name holding the main document body.
const documentXMLPath = "word/document.xml"

// trackedChangeElements are OOXML local element names (namespace-prefix
// stripped by encoding/xml's tokenizer) that indicate an unaccepted/unrejected
// tracked change. Presence of any means the buffer is not safe to freeze.
var trackedChangeElements = map[string]bool{
	"ins":       true,
	"del":       true,
	"pPrChange": true,
}

// unstrippedCommentElements are OOXML local element names indicating a
// comment anchor mark still present in the body. Comment marks must be
// stripped before freeze (spec.md §2.2 point 2); their presence means the
// document was not run through the strip step.
var unstrippedCommentElements = map[string]bool{
	"commentRangeStart": true,
	"commentReference":  true,
}

// ScanForUnresolvedMarkup unzips docxBytes, reads word/document.xml, and
// streams its XML tokens looking for any unresolved tracked-change or
// unstripped-comment element. It is a pure function with no I/O beyond the
// caller-supplied buffer — see spec.md Interview #2/#3: this feature does not
// yet wire a live docx-byte source into the freeze path (bounded defer), but
// the gate itself is fully built and tested per plan.md's literal task list
// so the future call site only needs to supply bytes.
//
// Matching is done on xml.Name.Local (the tokenizer already strips whichever
// namespace prefix alias the document declares for the wordprocessingml
// namespace, so this is robust to prefix aliasing, e.g. w: vs ns0:).
func ScanForUnresolvedMarkup(docxBytes []byte) error {
	if len(docxBytes) == 0 {
		return fmt.Errorf("%w: empty buffer", ErrInvalidDocxBuffer)
	}
	zr, err := zip.NewReader(bytes.NewReader(docxBytes), int64(len(docxBytes)))
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidDocxBuffer, err)
	}

	var docFile *zip.File
	for _, f := range zr.File {
		if f.Name == documentXMLPath {
			docFile = f
			break
		}
	}
	if docFile == nil {
		return fmt.Errorf("%w: no %s entry", ErrInvalidDocxBuffer, documentXMLPath)
	}

	rc, err := docFile.Open()
	if err != nil {
		return fmt.Errorf("%w: open %s: %v", ErrInvalidDocxBuffer, documentXMLPath, err)
	}
	defer rc.Close()

	decoder := xml.NewDecoder(rc)
	for {
		tok, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("%w: parse %s: %v", ErrInvalidDocxBuffer, documentXMLPath, err)
		}
		start, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		if trackedChangeElements[start.Name.Local] {
			return fmt.Errorf("%w: <%s>", ErrUnresolvedTrackedChanges, start.Name.Local)
		}
		if unstrippedCommentElements[start.Name.Local] {
			return fmt.Errorf("%w: <%s>", ErrUnstrippedComments, start.Name.Local)
		}
	}
	return nil
}
