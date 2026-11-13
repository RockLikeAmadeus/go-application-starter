package tokenstore

// Kept says where a token was kept.
type Kept string

const (
	InKeyring Kept = "keyring"
	InFile    Kept = "file"
)

// System keeps tokens in the operating system's credential store, and in a
// file when that store is unavailable.
type System struct {
	keyring *Keyring
	file    *File
}

// NewSystem returns a System whose file fallback lives in folder.
func NewSystem(folder string) *System {
	return &System{keyring: NewKeyring(), file: NewFile(folder)}
}

// Save keeps the token for the server and says where.
func (s *System) Save(serverURL, token string) (Kept, error) {
	if err := s.keyring.Save(serverURL, token); err == nil {
		return InKeyring, nil
	}
	return InFile, s.file.Save(serverURL, token)
}
