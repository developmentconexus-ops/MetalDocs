package application

import (
	"database/sql/driver"
	"io"
)

// docAreaRows is the shared fake-driver result for the ported B5 area read
// (LoadDocumentAreaCode): SELECT process_area_code_snapshot, controlled_document_id
// FROM documents ... — two columns. Returning a non-NULL snapshot (== the old
// single-value areaCode) and a NULL controlled_document_id reproduces the
// pre-port behavior exactly: the non-NULL snapshot wins the COALESCE and the
// NULL CD link means no CDFieldReader call. Used by every unit fake whose
// `from documents` branch served the area query before the port.
type docAreaRows struct {
	snapshot string
	done     bool
}

func (r *docAreaRows) Columns() []string {
	return []string{"process_area_code_snapshot", "controlled_document_id"}
}
func (r *docAreaRows) Close() error { return nil }
func (r *docAreaRows) Next(dest []driver.Value) error {
	if r.done {
		return io.EOF
	}
	r.done = true
	dest[0] = r.snapshot // non-NULL snapshot wins COALESCE → == old areaCode
	dest[1] = nil        // no CD link → no port call
	return nil
}
