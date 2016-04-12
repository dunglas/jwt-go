package jwt_test

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var (
	jwtTestDefaultKey *rsa.PublicKey
	defaultKeyFunc    jwt.Keyfunc = func(t *jwt.Token) (interface{}, error) { return jwtTestDefaultKey, nil }
	emptyKeyFunc      jwt.Keyfunc = func(t *jwt.Token) (interface{}, error) { return nil, nil }
	errorKeyFunc      jwt.Keyfunc = func(t *jwt.Token) (interface{}, error) { return nil, fmt.Errorf("error loading key") }
	nilKeyFunc        jwt.Keyfunc = nil
)

var jwtTestData = []struct {
	name        string
	tokenString string
	keyfunc     jwt.Keyfunc
	claims      jwt.Claims
	valid       bool
	errors      uint32
	parser      *jwt.Parser
}{
	{
		"basic",
		"eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJmb28iOiJiYXIifQ.FhkiHkoESI_cG3NPigFrxEk9Z60_oXrOT2vGm9Pn6RDgYNovYORQmmA0zs1AoAOf09ly2Nx2YAg6ABqAYga1AcMFkJljwxTT5fYphTuqpWdy4BELeSYJx5Ty2gmr8e7RonuUztrdD5WfPqLKMm1Ozp_T6zALpRmwTIW0QPnaBXaQD90FplAg46Iy1UlDKr-Eupy0i5SLch5Q-p2ZpaL_5fnTIUDlxC3pWhJTyx_71qDI-mAA_5lE_VdroOeflG56sSmDxopPEG3bFlSu1eowyBfxtu0_CuVd-M42RU75Zc4Gsj6uV77MBtbMrf4_7M_NUTSgoIF3fRqxrj0NzihIBg",
		defaultKeyFunc,
		jwt.MapClaims{"foo": "bar"},
		true,
		0,
		nil,
	},
	{
		"basic expired",
		"", // autogen
		defaultKeyFunc,
		jwt.MapClaims{"foo": "bar", "exp": float64(time.Now().Unix() - 100)},
		false,
		jwt.ValidationErrorExpired,
		nil,
	},
	{
		"basic nbf",
		"", // autogen
		defaultKeyFunc,
		jwt.MapClaims{"foo": "bar", "nbf": float64(time.Now().Unix() + 100)},
		false,
		jwt.ValidationErrorNotValidYet,
		nil,
	},
	{
		"expired and nbf",
		"", // autogen
		defaultKeyFunc,
		jwt.MapClaims{"foo": "bar", "nbf": float64(time.Now().Unix() + 100), "exp": float64(time.Now().Unix() - 100)},
		false,
		jwt.ValidationErrorNotValidYet | jwt.ValidationErrorExpired,
		nil,
	},
	{
		"basic invalid",
		"eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJmb28iOiJiYXIifQ.EhkiHkoESI_cG3NPigFrxEk9Z60_oXrOT2vGm9Pn6RDgYNovYORQmmA0zs1AoAOf09ly2Nx2YAg6ABqAYga1AcMFkJljwxTT5fYphTuqpWdy4BELeSYJx5Ty2gmr8e7RonuUztrdD5WfPqLKMm1Ozp_T6zALpRmwTIW0QPnaBXaQD90FplAg46Iy1UlDKr-Eupy0i5SLch5Q-p2ZpaL_5fnTIUDlxC3pWhJTyx_71qDI-mAA_5lE_VdroOeflG56sSmDxopPEG3bFlSu1eowyBfxtu0_CuVd-M42RU75Zc4Gsj6uV77MBtbMrf4_7M_NUTSgoIF3fRqxrj0NzihIBg",
		defaultKeyFunc,
		jwt.MapClaims{"foo": "bar"},
		false,
		jwt.ValidationErrorSignatureInvalid,
		nil,
	},
	{
		"basic nokeyfunc",
		"eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJmb28iOiJiYXIifQ.FhkiHkoESI_cG3NPigFrxEk9Z60_oXrOT2vGm9Pn6RDgYNovYORQmmA0zs1AoAOf09ly2Nx2YAg6ABqAYga1AcMFkJljwxTT5fYphTuqpWdy4BELeSYJx5Ty2gmr8e7RonuUztrdD5WfPqLKMm1Ozp_T6zALpRmwTIW0QPnaBXaQD90FplAg46Iy1UlDKr-Eupy0i5SLch5Q-p2ZpaL_5fnTIUDlxC3pWhJTyx_71qDI-mAA_5lE_VdroOeflG56sSmDxopPEG3bFlSu1eowyBfxtu0_CuVd-M42RU75Zc4Gsj6uV77MBtbMrf4_7M_NUTSgoIF3fRqxrj0NzihIBg",
		nilKeyFunc,
		jwt.MapClaims{"foo": "bar"},
		false,
		jwt.ValidationErrorUnverifiable,
		nil,
	},
	{
		"basic nokey",
		"eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJmb28iOiJiYXIifQ.FhkiHkoESI_cG3NPigFrxEk9Z60_oXrOT2vGm9Pn6RDgYNovYORQmmA0zs1AoAOf09ly2Nx2YAg6ABqAYga1AcMFkJljwxTT5fYphTuqpWdy4BELeSYJx5Ty2gmr8e7RonuUztrdD5WfPqLKMm1Ozp_T6zALpRmwTIW0QPnaBXaQD90FplAg46Iy1UlDKr-Eupy0i5SLch5Q-p2ZpaL_5fnTIUDlxC3pWhJTyx_71qDI-mAA_5lE_VdroOeflG56sSmDxopPEG3bFlSu1eowyBfxtu0_CuVd-M42RU75Zc4Gsj6uV77MBtbMrf4_7M_NUTSgoIF3fRqxrj0NzihIBg",
		emptyKeyFunc,
		jwt.MapClaims{"foo": "bar"},
		false,
		jwt.ValidationErrorSignatureInvalid,
		nil,
	},
	{
		"basic errorkey",
		"eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJmb28iOiJiYXIifQ.FhkiHkoESI_cG3NPigFrxEk9Z60_oXrOT2vGm9Pn6RDgYNovYORQmmA0zs1AoAOf09ly2Nx2YAg6ABqAYga1AcMFkJljwxTT5fYphTuqpWdy4BELeSYJx5Ty2gmr8e7RonuUztrdD5WfPqLKMm1Ozp_T6zALpRmwTIW0QPnaBXaQD90FplAg46Iy1UlDKr-Eupy0i5SLch5Q-p2ZpaL_5fnTIUDlxC3pWhJTyx_71qDI-mAA_5lE_VdroOeflG56sSmDxopPEG3bFlSu1eowyBfxtu0_CuVd-M42RU75Zc4Gsj6uV77MBtbMrf4_7M_NUTSgoIF3fRqxrj0NzihIBg",
		errorKeyFunc,
		jwt.MapClaims{"foo": "bar"},
		false,
		jwt.ValidationErrorUnverifiable,
		nil,
	},
	{
		"invalid signing method",
		"",
		defaultKeyFunc,
		jwt.MapClaims{"foo": "bar"},
		false,
		jwt.ValidationErrorSignatureInvalid,
		&jwt.Parser{ValidMethods: []string{"HS256"}},
	},
	{
		"valid signing method",
		"",
		defaultKeyFunc,
		jwt.MapClaims{"foo": "bar"},
		true,
		0,
		&jwt.Parser{ValidMethods: []string{"RS256", "HS256"}},
	},
	{
		"JSON Number",
		"",
		defaultKeyFunc,
		jwt.MapClaims{"foo": json.Number("123.4")},
		true,
		0,
		&jwt.Parser{UseJSONNumber: true},
	},
	{
		"Standard Claims",
		"",
		defaultKeyFunc,
		&jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Second * 10).Unix(),
		},
		true,
		0,
		&jwt.Parser{UseJSONNumber: true},
	},
}

func init() {
	if keyData, e := ioutil.ReadFile("test/sample_key.pub"); e == nil {
		if jwtTestDefaultKey, e = jwt.ParseRSAPublicKeyFromPEM(keyData); e != nil {
			panic(e)
		}
	} else {
		panic(e)
	}
}

func makeSample(c jwt.Claims) string {
	keyData, e := ioutil.ReadFile("test/sample_key")
	if e != nil {
		panic(e.Error())
	}
	key, e := jwt.ParseRSAPrivateKeyFromPEM(keyData)
	if e != nil {
		panic(e.Error())
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, c)
	s, e := token.SignedString(key)

	if e != nil {
		panic(e.Error())
	}

	return s
}

func TestParser_Parse(t *testing.T) {
	// Iterate over test data set and run tests
	for _, data := range jwtTestData {
		// If the token string is blank, use helper function to generate string
		if data.tokenString == "" {
			data.tokenString = makeSample(data.claims)
		}

		// Parse the token
		var token *jwt.Token
		var err error
		var parser = data.parser
		if parser == nil {
			parser = new(jwt.Parser)
		}
		// Figure out correct claims type
		switch data.claims.(type) {
		case jwt.MapClaims:
			token, err = parser.ParseWithClaims(data.tokenString, data.keyfunc, jwt.MapClaims{})
		case *jwt.StandardClaims:
			token, err = parser.ParseWithClaims(data.tokenString, data.keyfunc, &jwt.StandardClaims{})
		}

		// Verify result matches expectation
		if !reflect.DeepEqual(data.claims, token.Claims) {
			t.Errorf("[%v] Claims mismatch. Expecting: %v  Got: %v", data.name, data.claims, token.Claims)
		}

		if data.valid && err != nil {
			t.Errorf("[%v] Error while verifying token: %T:%v", data.name, err, err)
		}

		if !data.valid && err == nil {
			t.Errorf("[%v] Invalid token passed validation", data.name)
		}

		if data.errors != 0 {
			if err == nil {
				t.Errorf("[%v] Expecting error.  Didn't get one.", data.name)
			} else {
				// compare the bitfield part of the error
				if e := err.(*jwt.ValidationError).Errors; e != data.errors {
					t.Errorf("[%v] Errors don't match expectation.  %v != %v", data.name, e, data.errors)
				}
			}
		}
		if data.valid && token.Signature == "" {
			t.Errorf("[%v] Signature is left unpopulated after parsing", data.name)
		}
	}
}

func TestParseRequest(t *testing.T) {
	// Bearer token request
	for _, data := range jwtTestData {
		// FIXME: custom parsers are not supported by this helper.  skip tests that require them
		if data.parser != nil {
			t.Logf("Skipping [%v].  Custom parsers are not supported by ParseRequest", data.name)
			continue
		}

		if data.tokenString == "" {
			data.tokenString = makeSample(data.claims)
		}

		r, _ := http.NewRequest("GET", "/", nil)
		r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", data.tokenString))
		token, err := jwt.ParseFromRequestWithClaims(r, data.keyfunc, jwt.MapClaims{})

		if token == nil {
			t.Errorf("[%v] Token was not found: %v", data.name, err)
			continue
		}
		if !reflect.DeepEqual(data.claims, token.Claims) {
			t.Errorf("[%v] Claims mismatch. Expecting: %v  Got: %v", data.name, data.claims, token.Claims)
		}
		if data.valid && err != nil {
			t.Errorf("[%v] Error while verifying token: %v", data.name, err)
		}
		if !data.valid && err == nil {
			t.Errorf("[%v] Invalid token passed validation", data.name)
		}
	}
}

// Helper method for benchmarking various methods
func benchmarkSigning(b *testing.B, method jwt.SigningMethod, key interface{}) {
	t := jwt.New(method)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, err := t.SignedString(key); err != nil {
				b.Fatal(err)
			}
		}
	})

}