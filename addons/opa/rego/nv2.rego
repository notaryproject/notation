package kubernetes.admission
import data.kubernetes.namespaces


operations := {"CREATE", "UPDATE"}
kind := {"Deployment", "Pod"}

artifactType := "application/vnd.cncf.notary.v2"
mediaType := "application/vnd.cncf.notary.signature.v2+jwt"


## tbd: how to load certs
trusted_certificates := [
  {"registry": "invalid-registry.wabbit-networks.io", "cert": "abcdefghijklmnopqrstuvwxyz"},
  {"registry": "registry.wabbit-networks.io", "cert": "-----BEGIN CERTIFICATE-----\nMIID+TCCAuGgAwIBAgIUGJhuD3DXrW0ZZF8b7yu4tVB5hz4wDQYJKoZIhvcNAQEL\nBQAweDEkMCIGA1UEAwwbcmVnaXN0cnkud2FiYml0LW5ldHdvcmtzLmlvMRwwGgYD\nVQQKDBN3YWJiaXQtbmV0d29ya3MgaW5jMQswCQYDVQQGEwJVUzETMBEGA1UECAwK\nV2FzaGluZ3RvbjEQMA4GA1UEBwwHU2VhdHRsZTAeFw0yMTA1MTExNzE1NTBaFw0y\nMjA1MTExNzE1NTBaMHgxJDAiBgNVBAMMG3JlZ2lzdHJ5LndhYmJpdC1uZXR3b3Jr\ncy5pbzEcMBoGA1UECgwTd2FiYml0LW5ldHdvcmtzIGluYzELMAkGA1UEBhMCVVMx\nEzARBgNVBAgMCldhc2hpbmd0b24xEDAOBgNVBAcMB1NlYXR0bGUwggEiMA0GCSqG\nSIb3DQEBAQUAA4IBDwAwggEKAoIBAQC8kdL4Urb4AReee5mH/06l6/Vf1f8yPnqI\nkG2awNr6pdCe0ZwNbPKmrczLfypvYC4/Key1TAq0lHOz41gZNies0dw4IheiCpcZ\nRLMfIU3gvhVi2ohAuxQXJ03ZVpAAREPXr4F5YAh/Ny3LTTrn6NUpv+BqQ+DK/fbt\ndbRILBLBBMIaULbnKXWXduCii1al7jvBUgd4zW4zpojKC8EFmtMy1Ql2PNr6eE2Q\nNyilSKMkdasQopzwstW+BOoP1KCVhC/JxBcuCSU+/3o6Wbi4yAi/htt+sLcOyg/g\nrORbzS0XyuNdWm4kkPUCB+tNfWm35QU5R2ded/uOh5aFbkjmPrSXAgMBAAGjezB5\nMB0GA1UdDgQWBBSHxoHieiHjQoJDtLcTpuhvBC4yWjAfBgNVHSMEGDAWgBSHxoHi\neiHjQoJDtLcTpuhvBC4yWjAPBgNVHRMBAf8EBTADAQH/MCYGA1UdEQQfMB2CG3Jl\nZ2lzdHJ5LndhYmJpdC1uZXR3b3Jrcy5pbzANBgkqhkiG9w0BAQsFAAOCAQEACthX\n6iCbmlq9ZNU4KELwNfNvuyMIuCITzVHPasymaq7O5zNYI7Lu+7J2aTeWQHGmVrRS\njse33cfPxpg12uIYWdjP8WwlptU9BuCTNIZre9Bvc2LlV7oACuFAI1CRQA45bsbM\ngFEYhXbzvk6gfVBnYmKmn97Iyf+T5PnRyHukd9TqYz7p62sTbg/KBx1mzwLwVZLW\n0qPBiI5KDsPrp0TE7gtENev0eKJcK4iVRsgfw36oBkTdZHNh+Uzs6egXMpbJBYDT\nIp7TrwRo/Ce0pbAJ1HocKRrGeNpB2J39r+Q0XZDv/Vl3I7XG8GFu2xB2iL2oqKZ1\nWBU1wqZRolLtGVtUxw==\n-----END CERTIFICATE-----"},
  {"registry": "registry.acme-rockets", "cert": "123456789"}
]


image = get_images[_]
image_obj := split_image(image)


########################################################################################################


# deny not reachable registry
deny[msg] {
  operations[input.request.operation]
	kind[input.request.kind.kind]

  ping_registry(image_obj)
  msg := sprintf("cannot reach reagistry: %v", [image_obj.registry])	
}

# deny tag latest
deny[msg] {

	operations[input.request.operation]
	kind[input.request.kind.kind]

	contains(image.name, ":latest")

	msg := sprintf("%v contains tag `latest`; images with tag `lastet are not allowed", [image.name])
}

# deny not verifiable signature
deny[msg] {

	operations[input.request.operation]
	kind[input.request.kind.kind]

  # check endpoint avail
	digest := get_references(image_obj)
	nv2_jwt := get_nv2_jwt(image_obj, digest)

  x := verify_true(trusted_certificates, nv2_jwt)
  count(x) < 1
  msg := sprintf("signature for image: %v could not be verified", [image_obj.image])
	
}


########################################################################################################


# get images if pod
get_images[x] {

  kind[input.request.kind.kind]
	name := input.request.object.spec.containers[i].image

	x := {
		"index": i,
		"name": name,
	}
}

## get images if deployment
get_images[x] {

  kind[input.request.kind.kind]
	name := input.request.object.spec.template.spec.containers[i].image

	x := {
		"index": i,
		"name": name,
	}
}


## split image with digest
split_image(image) = x {

  registry_name := split(image.name, "/")[0]

  index_reg := indexof(image.name, "/")
  len_reg := count(image.name)
  image_tag := substring(image.name, index_reg+1, len_reg)

  contains(image_tag, "@sha256:")

  index_tag := indexof(image_tag, "@")
  len_tag = count(image_tag)

  digest := substring(image_tag, index_tag, len_tag)
  image_name := substring(image_tag, 0, index_tag)

  x := {
    "registry": registry_name,
    "image": image_name,
    "tag_digest": digest
  }
}

## split image with tag
split_image(image) = x {

  registry_name := split(image.name, "/")[0]
  
	index_reg := indexof(image.name, "/")
  len_reg := count(image.name)
  image_tag := substring(image.name, index_reg+1, len_reg)
  
	not contains(image_tag, "@sha256:")
	contains(image_tag, ":")

  index_tag := indexof(image_tag, ":")
  len_tag = count(image_tag)

  tag := substring(image_tag, index_tag+1, len_tag)
  image_name := substring(image_tag, 0, index_tag)

  x := {
    "registry": registry_name,
    "image": image_name,
    "tag_digest": tag
  }
}


## get references
get_references(image_obj) = signature_blob_digest {

	url_references_arr := [ "https:/", image_obj.registry, "v2/_ext/oci-artifacts/v1", image_obj.image, "manifests", image_obj.tag_digest, "references?n=10" ]
  url_references := concat("/", url_references_arr)
  
	headers_json := {"Content-Type": "application/json"}
  response_references := http.send({"method": "get", "url": url_references, "headers": headers_json, "tls_insecure_skip_verify": true})
  response_references.status_code == 200
  
	artifactType == response_references.body.references[index].manifest.artifactType
  mediaType == response_references.body.references[index].manifest.blobs[0].mediaType
  
	signature_blob_digest := response_references.body.references[index].manifest.blobs[0].digest
}

## get nv2 signature
get_nv2_jwt(image_obj, digest) = nv2_jwt {
  
	url_digest_arr := [ "https:/", image_obj.registry, "v2", image_obj.image, "blobs", digest ]
  url_digest := concat("/", url_digest_arr)
  
	headers_json := {"Content-Type": "application/json"}
  response_signature := http.send({"method": "get", "url": url_digest, "headers": headers_json, "tls_insecure_skip_verify": true})
  response_signature.status_code == 200
  
	nv2_jwt := response_signature.raw_body
}


## verify nv2_jwt with trusted certs 
## get trusted certs with registry
verify_true(trusted_certificates, nv2_jwt) = result {

  result := [p |
		cert_obj = trusted_certificates[_]
    io.jwt.verify_rs256(nv2_jwt, cert_obj.cert)

		p := {
			"registry": cert_obj.registry,
			"cert": cert_obj.cert
		}
	]
}

# check if registry is reachable
ping_registry(image_obj) {

  url_v2_arr := [ "https:/", image_obj.registry, "v2/" ]
  url_v2 := concat("/", url_v2_arr)

  headers_json := {"Content-Type": "application/json"}
  not http.send({"method": "get", "url": url_v2, "headers": headers_json, "tls_insecure_skip_verify": true})
}
