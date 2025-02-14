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

import os
import argparse
import subprocess
import threading
from http.server import BaseHTTPRequestHandler
from socketserver import TCPServer

# Global variable to hold the OCSP server process
ocsp_process = None
ocsp_lock = threading.Lock()

def start_ocsp_server(config_dir):
    global ocsp_process
    # Start OCSP server in background in config_dir
    cmd = [
        "openssl", "ocsp",
        "-port", "10087",
        "-index", "demoCA/index.txt",
        "-CA", "root.crt",
        "-rkey", "ocsp.key",
        "-rsigner", "ocsp.crt",
        "-nmin", "5",
        "-text",
    ]
    ocsp_process = subprocess.Popen(cmd, cwd=config_dir)
    print("OCSP server started with PID:", ocsp_process.pid)

def stop_ocsp_server():
    global ocsp_process
    if ocsp_process:
        ocsp_process.terminate()
        ocsp_process.wait()
        print("OCSP server with PID", ocsp_process.pid, "terminated")
        ocsp_process = None

def restart_ocsp_server(config_dir):
    with ocsp_lock:
        stop_ocsp_server()
        start_ocsp_server(config_dir)

def update_index_file(config_dir, new_content):
    index_path = os.path.join(config_dir, "demoCA", "index.txt")
    with open(index_path, "w") as f:
        f.write(new_content + "\n")
    print("Updated", index_path)

class OCSPRequestHandler(BaseHTTPRequestHandler):
    def do_POST(self):
        response = ""
        # Ensure path is processed without query parameters.
        path = self.path.split("?")[0]
        if path == "/revoke":
            # Revoke: update index.txt with revoked entry
            revoke_content = ("R	21250121012109Z	250214013720Z	520D9B1364D98367711DA2B6A0A0F34B23E3D02A	unknown	/C=US/ST=State/L=City/O=Organization/OU=OrgUnit/CN=LeafCert")
            update_index_file(self.server.config_dir, revoke_content)
            restart_ocsp_server(self.server.config_dir)
            response = "OCSP server restarted with index updated (revoke)."
        elif path == "/unrevoke":
            # Unrevoke: update index.txt with valid entry
            unrevoke_content = ("V	21250121012109Z	250214013720Z	520D9B1364D98367711DA2B6A0A0F34B23E3D02A	unknown	/C=US/ST=State/L=City/O=Organization/OU=OrgUnit/CN=LeafCert")
            update_index_file(self.server.config_dir, unrevoke_content)
            restart_ocsp_server(self.server.config_dir)
            response = "OCSP server restarted with index updated (unrevoke)."
        elif path == "/unknown":
            # Unknown endpoint
            empty_content = ""
            update_index_file(self.server.config_dir, empty_content)
            restart_ocsp_server(self.server.config_dir)
            response = "OCSP server restarted with empty index."
        else:
            response = "Invalid endpoint. Use /revoke or /unrevoke."

        self.send_response(200)
        self.send_header("Content-type", "text/plain")
        self.end_headers()
        self.wfile.write(response.encode("utf-8"))

class ReusableTCPServer(TCPServer):
    allow_reuse_address = True

def run_server(config_dir, host="localhost", port=10088):
    server_address = (host, port)
    httpd = ReusableTCPServer(server_address, OCSPRequestHandler)
    httpd.config_dir = config_dir  # attach config directory to server instance
    print(f"HTTP control server running on {host}:{port}")
    httpd.serve_forever()

if __name__ == "__main__":
    parser = argparse.ArgumentParser(
        description="Start OCSP server control HTTP server."
    )
    parser.add_argument(
        "--config-dir",
        required=True,
        help="Path to OCSP configuration folder."
    )
    args = parser.parse_args()
    config_dir = os.path.abspath(args.config_dir)
    # Change to config_dir to ensure all commands run there (optional)
    os.chdir(config_dir)
    # Start the OCSP server in background
    start_ocsp_server(config_dir)
    # Run HTTP control server on port 10088
    run_server(config_dir)
