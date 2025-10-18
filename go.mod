module github.com/cmlabs-hris/hris-backend-go

go 1.25.3

require (
	// Direct dependencies
	github.com/go-chi/chi/v5 v5.2.3
	github.com/go-chi/cors v1.2.2
	github.com/go-chi/httplog/v3 v3.2.2
	github.com/go-chi/jwtauth/v5 v5.3.3
	github.com/jackc/pgx/v5 v5.7.6
	github.com/joho/godotenv v1.5.1
	github.com/lestrrat-go/jwx/v2 v2.1.6
	golang.org/x/crypto v0.43.0
)

require (
	// Indirect dependencies
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.4.0 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/lestrrat-go/blackmagic v1.0.4 // indirect
	github.com/lestrrat-go/httpcc v1.0.1 // indirect
	github.com/lestrrat-go/httprc v1.0.6 // indirect
	github.com/lestrrat-go/iter v1.0.2 // indirect
	github.com/lestrrat-go/option v1.0.1 // indirect
	github.com/segmentio/asm v1.2.1 // indirect
	golang.org/x/oauth2 v0.32.0
	golang.org/x/sync v0.17.0 // indirect
	golang.org/x/sys v0.37.0 // indirect
	golang.org/x/text v0.30.0 // indirect
)

require github.com/stretchr/testify v1.10.0

require (
	cloud.google.com/go/compute/metadata v0.3.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
