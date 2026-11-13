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

// Read returns the server's token from the keyring, else from the file, and
// ErrNoToken when neither has one.
func (s *System) Read(serverURL string) (string, error) {
	if token, err := s.keyring.Read(serverURL); err == nil {
		return token, nil
	}
	return s.file.Read(serverURL)
}
