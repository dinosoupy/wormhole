package sender

import "io"

// SDPProvider returns the underlying SDPProvider
func (s *SenderSession) SDPProvider() io.Reader {
	return s.session.SDPProvider()
}
