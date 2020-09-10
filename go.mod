module github.com/goswap/stats-api

go 1.15

require (
	cloud.google.com/go/firestore v1.3.0
	cloud.google.com/go/storage v1.11.0 // indirect
	github.com/aristanetworks/goarista v0.0.0-20200812190859-4cb0e71f3c0e // indirect
	github.com/btcsuite/btcd v0.21.0-beta // indirect
	github.com/go-chi/chi v4.1.2+incompatible
	github.com/google/uuid v1.1.2 // indirect
	github.com/goswap/collector v0.0.5
	github.com/pborman/uuid v1.2.1 // indirect
	github.com/treeder/firetils v0.0.8
	github.com/treeder/gcputils v0.0.34
	github.com/treeder/goapibase v0.0.6
	github.com/treeder/gotils v0.0.15
	go.uber.org/zap v1.16.0 // indirect
	golang.org/x/crypto v0.0.0-20200820211705-5c72a883971a // indirect
	golang.org/x/sys v0.0.0-20200909081042-eff7692f9009 // indirect
	golang.org/x/tools v0.0.0-20200909210914-44a2922940c2 // indirect
	google.golang.org/api v0.31.0
	google.golang.org/grpc v1.32.0 // indirect
)

// replace github.com/goswap/collector => ../collector
