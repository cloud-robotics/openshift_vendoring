package client

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/openshift/golang.org/x/net/context"

	distreference "github.com/openshift/github.com/docker/distribution/reference"
	"github.com/openshift/github.com/docker/engine-api/types"
	"github.com/openshift/github.com/docker/engine-api/types/reference"
)

// ImageTag tags an image in the docker host
func (cli *Client) ImageTag(ctx context.Context, imageID, ref string, options types.ImageTagOptions) error {
	distributionRef, err := distreference.ParseNamed(ref)
	if err != nil {
		return fmt.Errorf("Error parsing reference: %q is not a valid repository/tag", ref)
	}

	if _, isCanonical := distributionRef.(distreference.Canonical); isCanonical {
		return errors.New("refusing to create a tag with a digest reference")
	}

	tag := reference.GetTagFromNamedRef(distributionRef)

	query := url.Values{}
	query.Set("repo", distributionRef.Name())
	query.Set("tag", tag)
	if options.Force {
		query.Set("force", "1")
	}

	resp, err := cli.post(ctx, "/images/"+imageID+"/tag", query, nil, nil)
	ensureReaderClosed(resp)
	return err
}
