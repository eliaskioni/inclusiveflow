package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/nyaruka/goflow/assets"
	"github.com/nyaruka/goflow/utils"

	log "github.com/sirupsen/logrus"
)

const (
	assetTypeChannel           AssetType = "channel"
	assetTypeField             AssetType = "field"
	assetTypeFlow              AssetType = "flow"
	assetTypeGroup             AssetType = "group"
	assetTypeLabel             AssetType = "label"
	assetTypeLocationHierarchy AssetType = "location_hierarchy"
	assetTypeResthook          AssetType = "resthook"
)

type LegacyServer interface {
	GetAsset(AssetType, string) (interface{}, error)
}

// ServerSource is an asset source which fetches assets from a server and caches them
type ServerSource struct {
	authToken  string
	typeURLs   map[AssetType]string
	httpClient *utils.HTTPClient
	cache      *AssetCache

	fetcher assetFetcher
}

var _ assets.AssetSource = (*ServerSource)(nil)
var _ assetFetcher = (*ServerSource)(nil)

// NewServerSource creates a new server asset source
func NewServerSource(authToken string, typeURLs map[AssetType]string, httpClient *utils.HTTPClient, cache *AssetCache) *ServerSource {
	// TODO validate typeURLs are for registered types?

	s := &ServerSource{authToken: authToken, typeURLs: typeURLs, httpClient: httpClient, cache: cache}
	s.fetcher = s
	return s
}

type serverSourceEnvelope struct {
	TypeURLs map[AssetType]string `json:"type_urls"`
}

// ReadServerSource reads a server asset source fronm the given JSON
func ReadServerSource(authToken string, httpClient *utils.HTTPClient, cache *AssetCache, data json.RawMessage) (*ServerSource, error) {
	envelope := &serverSourceEnvelope{}
	if err := utils.UnmarshalAndValidate(data, envelope); err != nil {
		return nil, fmt.Errorf("unable to read asset server: %s", err)
	}

	return NewServerSource(authToken, envelope.TypeURLs, httpClient, cache), nil
}

func (s *ServerSource) Channels() ([]assets.Channel, error) {
	asset, err := s.GetAsset(assetTypeChannel, "")
	if err != nil {
		return nil, err
	}
	set, isType := asset.([]assets.Channel)
	if !isType {
		return nil, fmt.Errorf("asset cache contains asset with wrong type")
	}
	return set, nil
}

func (s *ServerSource) Groups() ([]assets.Group, error) {
	asset, err := s.GetAsset(assetTypeGroup, "")
	if err != nil {
		return nil, err
	}
	set, isType := asset.([]assets.Group)
	if !isType {
		return nil, fmt.Errorf("asset cache contains asset with wrong type")
	}
	return set, nil
}

func (s *ServerSource) Labels() ([]assets.Label, error) {
	asset, err := s.GetAsset(assetTypeLabel, "")
	if err != nil {
		return nil, err
	}
	set, isType := asset.([]assets.Label)
	if !isType {
		return nil, fmt.Errorf("asset cache contains asset with wrong type")
	}
	return set, nil
}

func (s *ServerSource) Locations() ([]assets.Location, error) {
	// TODO
	return nil, nil
}

func (s *ServerSource) Resthooks() ([]assets.Resthook, error) {
	asset, err := s.GetAsset(assetTypeResthook, "")
	if err != nil {
		return nil, err
	}
	set, isType := asset.([]assets.Resthook)
	if !isType {
		return nil, fmt.Errorf("asset cache contains asset with wrong type")
	}
	return set, nil
}

func (s *ServerSource) HasLocations() bool {
	_, hasTypeURL := s.typeURLs["location_hierarchy"]
	return hasTypeURL
}

func (s *ServerSource) GetAsset(itemType AssetType, itemUUID string) (interface{}, error) {
	url, err := s.getAssetURL(itemType, itemUUID)
	if err != nil {
		return nil, err
	}

	return s.cache.getAsset(url, s.fetcher, itemType)
}

// getAssetURL gets the URL for a set of the given asset type
func (s *ServerSource) getAssetURL(itemType AssetType, itemUUID string) (string, error) {
	url, found := s.typeURLs[itemType]
	if !found {
		return "", fmt.Errorf("asset type '%s' not supported by asset server", itemType)
	}

	if itemUUID != "" {
		url = fmt.Sprintf("%s%s/", url, itemUUID)
	}

	return url, nil
}

// fetches an asset by its URL and parses it as the provided type
func (s *ServerSource) fetchAsset(url string, itemType AssetType) ([]byte, error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// set request headers
	request.Header.Set("Authorization", fmt.Sprintf("Token %s", s.authToken))

	// make the actual request
	response, err := s.httpClient.Do(request)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	log.WithField("asset_type", string(itemType)).WithField("url", url).Debugf("asset requested")

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("request returned non-200 response (%d)", response.StatusCode)
	}

	if response.Header.Get("Content-Type") != "application/json" {
		return nil, fmt.Errorf("request returned non-JSON response")
	}

	return ioutil.ReadAll(response.Body)
}

type MockServerSource struct {
	ServerSource

	mockResponses  map[string]json.RawMessage
	mockedRequests []string
}

// NewMockServerSource creates a new mocked asset server for testing
func NewMockServerSource(typeURLs map[AssetType]string, cache *AssetCache) *MockServerSource {
	s := &MockServerSource{
		ServerSource:   ServerSource{typeURLs: typeURLs, cache: cache},
		mockResponses:  map[string]json.RawMessage{},
		mockedRequests: []string{},
	}
	s.ServerSource.fetcher = s
	return s
}

func (s *MockServerSource) MockResponse(url string, response json.RawMessage) {
	s.mockResponses[url] = response
}

func (s *MockServerSource) MockedRequests() []string {
	return s.mockedRequests
}

func (s *MockServerSource) fetchAsset(url string, itemType AssetType) ([]byte, error) {
	s.mockedRequests = append(s.mockedRequests, url)

	assetBuf, found := s.mockResponses[url]
	if !found {
		return []byte(`{"results":[]}`), nil
	}
	return assetBuf, nil
}

// MarshalJSON marshals this mock asset server into JSON
func (s *MockServerSource) MarshalJSON() ([]byte, error) {
	envelope := &serverSourceEnvelope{TypeURLs: s.typeURLs}
	return json.Marshal(envelope)
}

var _ LegacyServer = (*MockServerSource)(nil)
var _ assetFetcher = (*MockServerSource)(nil)
