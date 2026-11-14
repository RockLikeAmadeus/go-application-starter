package acceptance

import (
	"strings"
	"testing"
)

func TestPostgresScenariosRunOnlyWhenThereIsADatabaseServer(t *testing.T) {
	tests := []struct {
		name                 string
		serverURL            string
		wantPostgresSkipped  bool
		wantServerDriverTags string
	}{
		{
			name:                 "with a database server",
			serverURL:            "postgres://localhost/test_server",
			wantPostgresSkipped:  false,
			wantServerDriverTags: "@server",
		},
		{
			name:                 "without a database server",
			serverURL:            "",
			wantPostgresSkipped:  true,
			wantServerDriverTags: "@server && ~@postgres",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			drivers := driversFor(t, tt.serverURL)

			postgresDriver := driverNamed(t, drivers, "postgres")
			if isSkipped := postgresDriver.skipReason != ""; isSkipped != tt.wantPostgresSkipped {
				t.Errorf("postgres driver skipReason = %q, want skipped = %v", postgresDriver.skipReason, tt.wantPostgresSkipped)
			}
			if tt.wantPostgresSkipped && !strings.Contains(postgresDriver.skipReason, "STARTER_TEST_DATABASE_URL") {
				t.Errorf("skipReason = %q, want it to name STARTER_TEST_DATABASE_URL", postgresDriver.skipReason)
			}
			if want := driverNamed(t, drivers, "sqlite").tags; postgresDriver.tags != want {
				t.Errorf("postgres driver tags = %q, want the sqlite driver's %q", postgresDriver.tags, want)
			}
			if got := driverNamed(t, drivers, "server").tags; got != tt.wantServerDriverTags {
				t.Errorf("server driver tags = %q, want %q", got, tt.wantServerDriverTags)
			}
		})
	}
}

// driverNamed finds the driver with the given name, failing the test if there is none.
func driverNamed(t *testing.T, drivers []driver, name string) driver {
	t.Helper()
	for _, d := range drivers {
		if d.name == name {
			return d
		}
	}
	t.Fatalf("no driver named %q", name)
	return driver{}
}

func TestSignInProviderScenariosRunOnlyOnTheOIDCDriver(t *testing.T) {
	drivers := driversFor(t, "")
	oidc := driverNamed(t, drivers, "oidc")
	if oidc.startHosted == nil {
		t.Error("the oidc driver has no hosted edition")
	}
	if strings.Contains(oidc.tags, "~@sign-in") || strings.Contains(oidc.tags, "~@sign-in-provider") {
		t.Errorf("oidc driver tags = %q, want it to run @sign-in and @sign-in-provider", oidc.tags)
	}
	for _, d := range drivers {
		if d.name == "oidc" || d.name == "server" { // the server driver runs only @server
			continue
		}
		if !strings.Contains(d.tags, "~@sign-in-provider") && !strings.Contains(d.tags, "~@sign-in") {
			t.Errorf("%s driver tags = %q, want it to leave out @sign-in-provider", d.name, d.tags)
		}
	}
}

func TestCommandLineSignInScenariosRunOnlyOnTheHTTPAndOIDCDrivers(t *testing.T) {
	for _, d := range driversFor(t, "") {
		switch d.name {
		case "http", "oidc":
			if d.startHosted == nil {
				t.Errorf("%s driver has no hosted edition", d.name)
			}
			if strings.Contains(d.tags, "~@command-line-sign-in") {
				t.Errorf("%s driver tags = %q, want it to run @command-line-sign-in", d.name, d.tags)
			}
		case "server": // runs only @server
		default:
			if !strings.Contains(d.tags, "~@command-line-sign-in") {
				t.Errorf("%s driver tags = %q, want it to leave out @command-line-sign-in", d.name, d.tags)
			}
		}
	}
	if oidc := driverNamed(t, driversFor(t, ""), "oidc"); !strings.Contains(oidc.tags, "@command-line-sign-in") {
		t.Errorf("oidc driver tags = %q, want it to include @command-line-sign-in", oidc.tags)
	}
}
