package docker

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"errors"
	"io"
	"strings"

	"github.com/distribution/distribution/v3"
	"github.com/distribution/distribution/v3/manifest/schema2"
	"github.com/notaryproject/notation/internal/ioutil"
	"github.com/opencontainers/go-digest"
)

type manifestInTar struct {
	Config   string
	RepoTags []string
	Layers   []string
}

// GenerateSchema2FromDockerSave generate a docker schema2 manifest from `docker save`
func GenerateSchema2FromDockerSave(reader io.Reader) (distribution.Manifest, error) {
	items, descriptors, err := extractTar(reader)
	if err != nil {
		return nil, err
	}
	if len(items) != 1 {
		return nil, errors.New("unsupported number of images")
	}
	item := items[0]

	layers := make([]distribution.Descriptor, 0, len(item.Layers))
	for _, layer := range item.Layers {
		layers = append(layers, descriptors[layer])
	}

	manifest := schema2.Manifest{
		Versioned: schema2.SchemaVersion,
		Config:    descriptors[item.Config],
		Layers:    layers,
	}
	return schema2.FromStruct(manifest)
}

func extractTar(r io.Reader) ([]manifestInTar, map[string]distribution.Descriptor, error) {
	var manifests []manifestInTar
	descriptors := make(map[string]distribution.Descriptor)

	tr := tar.NewReader(r)
	for {
		file, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, nil, err
		}
		switch {
		case file.Name == "manifest.json":
			decoder := json.NewDecoder(tr)
			if err := decoder.Decode(&manifests); err != nil {
				return nil, nil, err
			}
		case strings.HasSuffix(file.Name, "/layer.tar"):
			desc, err := generateLayerDescriptor(tr)
			if err != nil {
				return nil, nil, err
			}
			descriptors[file.Name] = desc
		case strings.HasSuffix(file.Name, ".json"):
			digest := digest.NewDigestFromEncoded(digest.SHA256, strings.TrimSuffix(file.Name, ".json"))
			descriptors[file.Name] = distribution.Descriptor{
				MediaType: schema2.MediaTypeImageConfig,
				Size:      file.Size,
				Digest:    digest,
			}
		}
	}

	return manifests, descriptors, nil
}

func generateLayerDescriptor(r io.Reader) (distribution.Descriptor, error) {
	digester := digest.SHA256.Digester()
	count := ioutil.NewCountWriter(digester.Hash())
	w := gzip.NewWriter(count)
	_, err := io.Copy(w, r)
	if err != nil {
		return distribution.Descriptor{}, err
	}
	if err := w.Close(); err != nil {
		return distribution.Descriptor{}, err
	}
	return distribution.Descriptor{
		MediaType: schema2.MediaTypeLayer,
		Size:      count.N,
		Digest:    digester.Digest(),
	}, nil
}
