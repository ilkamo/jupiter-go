package jupiter_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	jupiter "github.com/ilkamo/jupiter-go"
	"github.com/ilkamo/jupiter-go/studio"
)

const (
	studioAPIKey      = "studio-secret"
	imageUploadPath   = "/upload/image"
	imageUploadQuery  = "X-Signature=a%2Fb%3D&X-Expires=60"
	metadataUploadURI = "/upload/metadata"
	launchedMint      = "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"
)

type studioRequest struct {
	method   string
	path     string
	rawQuery string
	header   http.Header
	body     []byte
}

// studioServer serves Studio API and presigned upload paths over TLS, as
// presigned uploads require HTTPS.
func studioServer(
	t *testing.T,
	replies map[string]fakeRoute,
) (*jupiter.Client, string, func() []studioRequest) {
	t.Helper()

	var (
		mu       sync.Mutex
		requests []studioRequest
	)

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read body: %v", err)
		}

		mu.Lock()
		requests = append(requests, studioRequest{
			method:   r.Method,
			path:     r.URL.Path,
			rawQuery: r.URL.RawQuery,
			header:   r.Header.Clone(),
			body:     body,
		})
		mu.Unlock()

		reply, ok := replies[r.URL.Path]
		if !ok {
			reply = fakeRoute{status: http.StatusOK}
		}

		w.WriteHeader(reply.status)

		if _, err := w.Write([]byte(reply.body)); err != nil {
			t.Errorf("write response: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	client, err := jupiter.NewClient(
		jupiter.WithBaseURL(server.URL),
		jupiter.WithHTTPClient(server.Client()),
		jupiter.WithAPIKey(studioAPIKey),
	)
	if err != nil {
		t.Fatal(err)
	}

	return client, server.URL, func() []studioRequest {
		mu.Lock()
		defer mu.Unlock()

		return append([]studioRequest(nil), requests...)
	}
}

func memePool() studio.PoolRequest {
	return studio.PoolRequest{
		Curve: studio.CurveByMarketCap{
			QuoteMint:          usdcMint,
			InitialMarketCap:   "16000",
			MigrationMarketCap: "69000.5",
			TokenQuoteDecimal:  studio.QuoteDecimals6,
			LockedVesting:      studio.NoLockedVesting(),
		},
		AntiSniping:           true,
		Fee:                   studio.Fee{FeeBps: studio.FeeBps100},
		TokenName:             "My Token",
		TokenSymbol:           "MTK",
		TokenImageContentType: studio.ImagePNG,
		Creator:               takerAddress,
	}
}

func TestStudioLaunchFlow(t *testing.T) {
	var baseURL string
	createReply := func() string {
		return `{"transaction": "AQ==", "mint": "` + launchedMint + `",
"imagePresignedUrl": "` + baseURL + imageUploadPath + "?" + imageUploadQuery + `",
"metadataPresignedUrl": "` + baseURL + metadataUploadURI + `",
"imageUrl": "https://static-create.jup.ag/images/mint"}`
	}

	replies := map[string]fakeRoute{
		"/studio/v1/dbc-pool/submit": {
			status: http.StatusOK,
			body:   `{"success": true, "data": {"mint": "` + launchedMint + `", "configKey": "config"}}`,
		},
	}
	client, url, recorded := studioServer(t, replies)
	baseURL = url
	replies["/studio/v1/dbc-pool/create-tx"] = fakeRoute{status: http.StatusOK, body: createReply()}

	ctx := context.Background()
	pool, err := client.Studio.CreatePool(ctx, memePool())
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.Studio.UploadImage(ctx, pool.ImagePresignedURL, studio.ImagePNG, []byte("png"))
	if err != nil {
		t.Fatal(err)
	}

	metadata := studio.TokenMetadata{Name: "My Token", Symbol: "MTK", Image: pool.ImageURL}
	if _, err := client.Studio.UploadMetadata(ctx, pool.MetadataPresignedURL, metadata); err != nil {
		t.Fatal(err)
	}

	submitted, err := client.Studio.Submit(ctx, studio.SubmitRequest{
		SignedTransaction: "AQ==",
		Owner:             takerAddress,
		Content:           "lore",
		HeaderImage: &studio.HeaderImage{
			Filename:    "header.jpeg",
			ContentType: studio.ImageJPEG,
			Data:        []byte("jpeg"),
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if submitted.Mint != launchedMint || submitted.ConfigKey != "config" {
		t.Fatalf("submitted = %+v", submitted)
	}

	requests := recorded()
	create, image, meta, submit := requests[0], requests[1], requests[2], requests[3]

	var createBody map[string]json.RawMessage
	if err := json.Unmarshal(create.body, &createBody); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(createBody["buildCurveByMarketCapParam"]), `"migrationMarketCap":69000.5`) {
		t.Fatalf("curve = %s", createBody["buildCurveByMarketCapParam"])
	}

	if _, ok := createBody["isLpLocked"]; ok {
		t.Fatal("unset LP lock was sent instead of the upstream default")
	}

	for _, upload := range []studioRequest{image, meta} {
		if upload.method != http.MethodPut || upload.header.Get("x-api-key") != "" {
			t.Fatalf("upload %s carried credentials or used %s", upload.path, upload.method)
		}
	}

	imageKept := string(image.body) == "png" && image.header.Get("Content-Type") == studio.ImagePNG
	if image.rawQuery != imageUploadQuery || !imageKept {
		t.Fatalf("image upload = %+v", image)
	}

	metadataKept := strings.Contains(string(meta.body), `"symbol":"MTK"`)
	if meta.header.Get("Content-Type") != "application/json" || !metadataKept {
		t.Fatalf("metadata upload = %+v", meta)
	}

	for _, apiRequest := range []studioRequest{create, submit} {
		if apiRequest.header.Get("x-api-key") != studioAPIKey {
			t.Fatalf("API request %s lacks the key", apiRequest.path)
		}
	}

	assertSubmitForm(t, submit)
}

func assertSubmitForm(t *testing.T, submit studioRequest) {
	t.Helper()

	mediaType, params, err := mime.ParseMediaType(submit.header.Get("Content-Type"))
	if err != nil || mediaType != "multipart/form-data" {
		t.Fatalf("submit content type = %q (%v)", submit.header.Get("Content-Type"), err)
	}

	reader := multipart.NewReader(strings.NewReader(string(submit.body)), params["boundary"])
	form, err := reader.ReadForm(int64(len(submit.body)))
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := form.RemoveAll(); err != nil {
			t.Errorf("remove form files: %v", err)
		}
	}()

	values := form.Value
	ownerKept := values["owner"][0] == takerAddress
	if values["transaction"][0] != "AQ==" || !ownerKept || values["content"][0] != "lore" {
		t.Fatalf("form values = %v", form.Value)
	}

	header := form.File["headerImage"][0]
	if header.Filename != "header.jpeg" || header.Header.Get("Content-Type") != studio.ImageJPEG {
		t.Fatalf("header image = %+v", header)
	}
}

func TestStudioSubmitOutcomes(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		body    string
		unknown bool
	}{
		{name: "server error", status: http.StatusInternalServerError, body: `{}`, unknown: true},
		{name: "rejected", status: http.StatusBadRequest, body: `{"error": "invalid"}`},
		{name: "explicit failure", status: http.StatusOK, body: `{"success": false}`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client, _, recorded := studioServer(t, map[string]fakeRoute{
				"/studio/v1/dbc-pool/submit": {status: test.status, body: test.body},
			})

			request := studio.SubmitRequest{SignedTransaction: "AQ==", Owner: takerAddress}
			_, err := client.Studio.Submit(context.Background(), request)
			if err == nil {
				t.Fatal("expected failure")
			}

			if errors.Is(err, studio.ErrOutcomeUnknown) != test.unknown {
				t.Fatalf("unexpected outcome classification: %v", err)
			}

			if len(recorded()) != singleAttempt {
				t.Fatalf("submit attempts = %d", len(recorded()))
			}
		})
	}
}

func TestStudioPoolsAndFees(t *testing.T) {
	client, _, recorded := studioServer(t, map[string]fakeRoute{
		"/studio/v1/dbc-pool/addresses/" + launchedMint: {
			status: http.StatusOK,
			body:   `{"success": true, "data": {"dammv2PoolAddress": null, "dbcPoolAddress": "dbc"}}`,
		},
		"/studio/v1/dbc/fee": {
			status: http.StatusOK,
			body:   `{"unclaimed": "5000000", "total": "25000000"}`,
		},
		"/studio/v1/dbc/fee/create-tx": {status: http.StatusOK, body: `{"transaction": "AQ=="}`},
	})

	ctx := context.Background()
	pools, err := client.Studio.PoolAddresses(ctx, launchedMint)
	if err != nil || pools.DBCPoolAddress != "dbc" || pools.DAMMv2PoolAddress != "" {
		t.Fatalf("pools=%+v err=%v", pools, err)
	}

	fees, err := client.Studio.Fee(ctx, payerAddress)
	if err != nil || fees.Unclaimed != "5000000" {
		t.Fatalf("fees=%+v err=%v", fees, err)
	}

	claim := studio.ClaimFeeRequest{OwnerWallet: takerAddress, PoolAddress: payerAddress, MaxQuoteAmount: "0"}
	if _, err := client.Studio.ClaimFeeTransaction(ctx, claim); err != nil {
		t.Fatal(err)
	}

	sent := recorded()[2].body
	if !strings.Contains(string(sent), `"maxQuoteAmount":0`) {
		t.Fatalf("claim body = %s", sent)
	}
}

func TestStudioFeeReadRetries(t *testing.T) {
	server, calls := flakyServer(t, http.StatusServiceUnavailable, `{"unclaimed": "1", "total": "2"}`)

	client, err := jupiter.NewClient(jupiter.WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	fees, err := client.Studio.Fee(context.Background(), payerAddress)
	if err != nil {
		t.Fatal(err)
	}

	if calls.Load() != recoveredAttempts || fees.Metadata.Attempts != recoveredAttempts {
		t.Fatalf("calls=%d attempts=%d", calls.Load(), fees.Metadata.Attempts)
	}
}

func TestInvalidStudioRequestsDoNotReachTransport(t *testing.T) {
	client, err := jupiter.NewClient(jupiter.WithHTTPClient(unreachableHTTPClient{t: t}))
	if err != nil {
		t.Fatal(err)
	}

	s := client.Studio
	ctx := context.Background()

	pool := func(configure func(*studio.PoolRequest)) func() error {
		return func() error {
			request := memePool()
			configure(&request)
			_, err := s.CreatePool(ctx, request)

			return err
		}
	}

	tests := []struct {
		name string
		call func() error
		want error
	}{
		{
			name: "malformed quote mint",
			call: pool(func(r *studio.PoolRequest) { r.Curve.QuoteMint = "USDC" }),
			want: studio.ErrInvalidAddress,
		},
		{
			name: "malformed creator",
			call: pool(func(r *studio.PoolRequest) { r.Creator = "me" }),
			want: studio.ErrInvalidAddress,
		},
		{
			name: "migration not above initial",
			call: pool(func(r *studio.PoolRequest) { r.Curve.MigrationMarketCap = "16000" }),
		},
		{
			name: "zero initial market cap",
			call: pool(func(r *studio.PoolRequest) { r.Curve.InitialMarketCap = "0" }),
		},
		{
			name: "negative vesting amount",
			call: pool(func(r *studio.PoolRequest) { r.Curve.LockedVesting.CliffUnlockAmount = "-1" }),
		},
		{
			name: "missing vesting field",
			call: pool(func(r *studio.PoolRequest) { r.Curve.LockedVesting = studio.LockedVesting{} }),
		},
		{
			name: "missing token name",
			call: pool(func(r *studio.PoolRequest) { r.TokenName = "" }),
		},
		{
			name: "missing fee",
			call: pool(func(r *studio.PoolRequest) { r.Fee = studio.Fee{} }),
		},
		{
			name: "plain HTTP upload",
			call: func() error {
				plainURL := "http://uploads.example.com/image"
				_, err := s.UploadImage(ctx, plainURL, studio.ImagePNG, []byte("png"))
				return err
			},
		},
		{
			name: "empty image",
			call: func() error {
				_, err := s.UploadImage(ctx, "https://uploads.example.com/image", studio.ImagePNG, nil)
				return err
			},
		},
		{
			name: "submit without transaction",
			call: func() error {
				_, err := s.Submit(ctx, studio.SubmitRequest{Owner: takerAddress})
				return err
			},
		},
		{
			name: "header image without content type",
			call: func() error {
				_, err := s.Submit(ctx, studio.SubmitRequest{
					SignedTransaction: "AQ==",
					Owner:             takerAddress,
					HeaderImage:       &studio.HeaderImage{Filename: "h.png", Data: []byte("png")},
				})
				return err
			},
		},
		{
			name: "negative claim amount",
			call: func() error {
				_, err := s.ClaimFeeTransaction(ctx, studio.ClaimFeeRequest{
					OwnerWallet:    takerAddress,
					PoolAddress:    payerAddress,
					MaxQuoteAmount: "-1",
				})
				return err
			},
		},
		{
			name: "malformed pool mint",
			call: func() error {
				_, err := s.PoolAddresses(ctx, "mint")
				return err
			},
			want: studio.ErrInvalidAddress,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.call()
			if err == nil {
				t.Fatal("invalid request accepted")
			}

			if errors.Is(err, studio.ErrOutcomeUnknown) {
				t.Fatalf("validation failure reported as unknown outcome: %v", err)
			}

			if test.want != nil && !errors.Is(err, test.want) {
				t.Fatalf("expected %v, got %v", test.want, err)
			}
		})
	}
}
