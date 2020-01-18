package ccv3

import (
	"bytes"
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

// Relationship represents a one to one relationship.
// An empty GUID will be marshaled as `null`.
type Relationship struct {
	GUID string
}

func (r Relationship) MarshalJSON() ([]byte, error) {
	if r.GUID == "" {
		var emptyCCRelationship struct {
			Data interface{} `json:"data"`
		}
		return json.Marshal(emptyCCRelationship)
	}

	var ccRelationship struct {
		Data struct {
			GUID string `json:"guid"`
		} `json:"data"`
	}

	ccRelationship.Data.GUID = r.GUID
	return json.Marshal(ccRelationship)
}

func (r *Relationship) UnmarshalJSON(data []byte) error {
	var ccRelationship struct {
		Data struct {
			GUID string `json:"guid"`
		} `json:"data"`
	}

	err := cloudcontroller.DecodeJSON(data, &ccRelationship)
	if err != nil {
		return err
	}

	r.GUID = ccRelationship.Data.GUID
	return nil
}

// DeleteIsolationSegmentOrganization will delete the relationship between
// the isolation segment and the organization provided.
func (client *Client) DeleteIsolationSegmentOrganization(isolationSegmentGUID string, orgGUID string) (Warnings, error) {
	_, warnings, err := client.requester.MakeRequest(client, requestParams{
		RequestName: internal.DeleteIsolationSegmentRelationshipOrganizationRequest,
		URIParams:   internal.Params{"isolation_segment_guid": isolationSegmentGUID, "organization_guid": orgGUID},
	})

	return warnings, err
}

// DeleteServiceInstanceRelationshipsSharedSpace will delete the sharing relationship
// between the service instance and the shared-to space provided.
func (client *Client) DeleteServiceInstanceRelationshipsSharedSpace(serviceInstanceGUID string, spaceGUID string) (Warnings, error) {
	_, warnings, err := client.requester.MakeRequest(client, requestParams{
		RequestName: internal.DeleteServiceInstanceRelationshipsSharedSpaceRequest,
		URIParams:   internal.Params{"service_instance_guid": serviceInstanceGUID, "space_guid": spaceGUID},
	})

	return warnings, err
}

// GetOrganizationDefaultIsolationSegment returns the relationship between an
// organization and it's default isolation segment.
func (client *Client) GetOrganizationDefaultIsolationSegment(orgGUID string) (Relationship, Warnings, error) {
	request, err := client.NewHTTPRequest(requestOptions{
		RequestName: internal.GetOrganizationRelationshipDefaultIsolationSegmentRequest,
		URIParams:   internal.Params{"organization_guid": orgGUID},
	})
	if err != nil {
		return Relationship{}, nil, err
	}

	var relationship Relationship
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &relationship,
	}

	err = client.Connection.Make(request, &response)
	return relationship, response.Warnings, err
}

// GetSpaceIsolationSegment returns the relationship between a space and it's
// isolation segment.
func (client *Client) GetSpaceIsolationSegment(spaceGUID string) (Relationship, Warnings, error) {
	request, err := client.NewHTTPRequest(requestOptions{
		RequestName: internal.GetSpaceRelationshipIsolationSegmentRequest,
		URIParams:   internal.Params{"space_guid": spaceGUID},
	})
	if err != nil {
		return Relationship{}, nil, err
	}

	var relationship Relationship
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &relationship,
	}

	err = client.Connection.Make(request, &response)
	return relationship, response.Warnings, err
}

// SetApplicationDroplet sets the specified droplet on the given application.
func (client *Client) SetApplicationDroplet(appGUID string, dropletGUID string) (Relationship, Warnings, error) {
	relationship := Relationship{GUID: dropletGUID}
	bodyBytes, err := json.Marshal(relationship)
	if err != nil {
		return Relationship{}, nil, err
	}

	request, err := client.NewHTTPRequest(requestOptions{
		RequestName: internal.PatchApplicationCurrentDropletRequest,
		URIParams:   map[string]string{"app_guid": appGUID},
		Body:        bytes.NewReader(bodyBytes),
	})
	if err != nil {
		return Relationship{}, nil, err
	}

	var responseRelationship Relationship
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &responseRelationship,
	}
	err = client.Connection.Make(request, &response)

	return responseRelationship, response.Warnings, err
}

// UpdateOrganizationDefaultIsolationSegmentRelationship sets the default isolation segment
// for an organization on the controller.
// If isoSegGuid is empty it will reset the default isolation segment.
func (client *Client) UpdateOrganizationDefaultIsolationSegmentRelationship(orgGUID string, isoSegGUID string) (Relationship, Warnings, error) {
	body, err := json.Marshal(Relationship{GUID: isoSegGUID})
	if err != nil {
		return Relationship{}, nil, err
	}

	request, err := client.NewHTTPRequest(requestOptions{
		RequestName: internal.PatchOrganizationRelationshipDefaultIsolationSegmentRequest,
		Body:        bytes.NewReader(body),
		URIParams:   internal.Params{"organization_guid": orgGUID},
	})
	if err != nil {
		return Relationship{}, nil, err
	}

	var relationship Relationship
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &relationship,
	}
	err = client.Connection.Make(request, &response)
	return relationship, response.Warnings, err
}

// UpdateSpaceIsolationSegmentRelationship assigns an isolation segment to a space and
// returns the relationship.
func (client *Client) UpdateSpaceIsolationSegmentRelationship(spaceGUID string, isolationSegmentGUID string) (Relationship, Warnings, error) {
	body, err := json.Marshal(Relationship{GUID: isolationSegmentGUID})
	if err != nil {
		return Relationship{}, nil, err
	}

	request, err := client.NewHTTPRequest(requestOptions{
		RequestName: internal.PatchSpaceRelationshipIsolationSegmentRequest,
		URIParams:   internal.Params{"space_guid": spaceGUID},
		Body:        bytes.NewReader(body),
	})
	if err != nil {
		return Relationship{}, nil, err
	}

	var relationship Relationship
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &relationship,
	}

	err = client.Connection.Make(request, &response)
	return relationship, response.Warnings, err
}
