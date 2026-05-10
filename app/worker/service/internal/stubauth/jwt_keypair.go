package stubauth

import (
	"sync"

	"github.com/Servora-Kit/servora/security/jwt"
)

// DemoPrivateKeyPEM is a TEST-ONLY RSA-2048 private key (PKCS#8). Generated
// for the servora-example demo/authn lighthouse so the sign-token CLI and
// the master/worker servers share the same kid + pubkey across processes.
// DO NOT use in production.
const DemoPrivateKeyPEM = `-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQCcIMeSMwNTRdAv
rAXOKjjKw1uUm1TNVUaaUI5eJ7i4Cyhx+4y044uoX3Ewew6KCBq+Zw1kkZmpulNK
uBrH0zoMWybtfZKw5ycMYJPJRvauq5OAdtc3YlbiWoUMTgEM4jNH1yvWsXbL3g44
kmtqrmM/b5nB/JK5mhoqERWLrCOvH4xWf7t+ULVJ3lNinSqawrWJcL6ejB1BlcNx
7Q+d9NgTpHHkbV2+9ftFp6G/Xjknexfdkd1QHYsPumhSPV+qyKBU7vDlw4iAfcb5
3THvucfZR94GPmdhqlO3xFo3mVTrAvGFyOCWkftEpWCPLYZTcO7hvSRC+ImOwgMC
HEY2AGo9AgMBAAECggEAE7pMnVuWyvniUXPCFAffvzcqJj/uWjexQSR2qM0KGS9H
wAdSAzVRW0gcDE0wTB60CmUJGzEOopSpm6Jht+gnyJHn/FBGrdW3aGf3pJIz72Q8
shnSHexuIBHiB+j/VvpqEmTM5EbuRNpdH6bYrdU1MznNyuY4P/2J2tDn0QtCQwld
8wQsYU0d1N6Qn9PtF26eEvKkIRaaYA4qDt4QPe+lD5mwJWwjpsWC0N5GT+MFiKy6
UmcicLBT+PfRcpBWoIQd3lqYgUhnxmyTrPQ0S0P3x4TrJgYU08fOVf6PNxjLCGWP
MDL/Rer6rOhep3XOvdh7AKDI5e9p9qAXXluMoRbRfQKBgQDJDgBcZeMMXV0PVj7T
qkYlEMQx35htPxt7kvla2ix5Rpo0JaFn3SlK+TG4UPJvnLvX0tbEm68ybMVqUkKe
4UDds+z3bAURvhDA6bg2tcWb+9PlY17hu578my1ZPgPa3biUv/TpygVhFxbt0Qk8
z/G2zvFSIUrbi+SdPHBx3sLf6wKBgQDGy6mWgoKlfij9nR9Wzf7SJX4tGx2AArVS
wG+bSqI47Gz1N6GbFwdYdbppeQOnywpXOTYrqR+mOM3nY87RdFa/3O+VwBORmSUF
hWfPYCfT1mINIlM1FzOOVJwJTSzrz2HO4mMND1DjgoHRYE2VDDFCHf5hTbiwNHMN
l1Vc7j38dwKBgCGBrdm4OTCUVq/5pZrM48fFlYziQJrkS4Y6pkfX2FWVyJksNEwE
9Z7DDOA0zVKAgmWjg5tcfsQekH/5mZS04YSROcq6O9YLIOulh8fGX1pxi4zNFMD1
7bcXfWVECoxtKxfPLdfQjTjzCiU0EyAJX7Uho+IWHk2ccMsriWnQwBVlAoGAbak7
S8OCvjfx9LUP7JqFzvbPu6IRi+PykkuFRWzOQAhrsnmVtC/n5WxMAJK46X6fna35
q+wHgXIkY1gzZmd+0yfVIg5qvQ511a3ZrhOk5L6GKCifLdI2pnUV/iuMdChaE/3e
Ff406Mu9QPqW0XmAUrCo+pQdJVZJgV3RwQnLN9ECgYEAjyHaJcXxebecC6ndoE68
8vB9kgq9zfL63YArOrZMt0AGhjTEnvLvHxoctg9YQRuIhP3p+v1xYo0hqVSNNpxL
gaYiBkjsOnB946e63ABanzXhwfqBJMqVzEhessfTAlEabeELipYgb6frmc2rzxkK
3SZ4W5PqUtGNNmyWddA1ork=
-----END PRIVATE KEY-----
`

var (
	once     sync.Once
	signer   *jwt.Signer
	verifier *jwt.Verifier
)

// SharedKeypair returns a process-wide RSA Signer + Verifier sharing the
// SAME RSA key sourced from the baked DemoPrivateKeyPEM constant. The
// sign-token CLI and the master/worker servers BOTH call this so the
// verifier and the test signer share kid + pubkey.
func SharedKeypair() (*jwt.Signer, *jwt.Verifier) {
	once.Do(func() {
		s, err := jwt.NewSigner([]byte(DemoPrivateKeyPEM))
		if err != nil {
			panic(err)
		}
		signer = s
		v := jwt.NewVerifier()
		v.AddKey(s.KID(), s.PublicKey())
		verifier = v
	})
	return signer, verifier
}
