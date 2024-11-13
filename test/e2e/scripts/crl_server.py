# Copyright The Notary Project Authors.
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import http.server
import socketserver
import os

PORT = 10086
DATA_DIR = './testdata/config/crl'
leaf_crl = 'leaf.crl'
intermediate_crl = 'intermediate.crl'


class CRLRequestHandler(http.server.SimpleHTTPRequestHandler):
    def do_GET(self):
        global leaf_crl
        global intermediate_crl
        if self.path == '/leaf.crl':
            file_path = os.path.join(DATA_DIR, leaf_crl)
            self.crl_response(file_path)
        elif self.path == '/intermediate.crl':
            file_path = os.path.join(DATA_DIR, intermediate_crl)
            self.crl_response(file_path)
        else:
            self.send_error(404, 'Not Found')
    
    def crl_response(self, file_path):
        if os.path.exists(file_path):
            self.send_response(200)
            self.send_header('Content-Type', 'application/pkix-crl')
            self.end_headers()
            with open(file_path, 'rb') as f:
                self.wfile.write(f.read())
        else:
            self.send_error(404, 'File Not Found')

    def do_POST(self):
        global leaf_crl
        global intermediate_crl
        if self.path == '/leaf/revoke':
            leaf_crl = 'leaf_revoked.crl'
            self.post_response()
        elif self.path == '/leaf/unrevoke':
            leaf_crl = 'leaf.crl'
            self.post_response()
        elif self.path == '/leaf/expired':
            leaf_crl = 'leaf_expired.crl'
            self.post_response()
        elif self.path == '/intermediate/revoke':
            intermediate_crl = 'intermediate_revoked.crl'
            self.post_response()
        elif self.path == '/intermediate/unrevoke':
            intermediate_crl = 'intermediate.crl'
            self.post_response()
        else:
            self.send_error(404, 'Not Found')
    
    def post_response(self):
        self.send_response(201)
        self.end_headers()
        self.wfile.write(b'ok')

class ReusableTCPServer(socketserver.TCPServer):
    allow_reuse_address = True

with ReusableTCPServer(('', PORT), CRLRequestHandler) as httpd:
    print(f"Serving at port {PORT}")
    try:
        httpd.serve_forever()
    finally:
        httpd.server_close()