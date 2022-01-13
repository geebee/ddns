/*
   Copyright 2022 https://github.com/geebee

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

addEventListener("fetch", (event) => {
  event.respondWith(
    handleRequest(event.request).catch(
      (err) => new Response(err.stack, { status: 500 })
    )
  );
});

async function handleRequest(request) {
  const remoteIP = request.headers.get("CF-Connecting-IP");
  const { pathname } = new URL(request.url);

  if (request.method != "GET") {
      return new Response("405 Method Not Allowed", {
        status: 405,
        headers: {"Content-Type": "text/plain"},
      });
  }

  switch (pathname) {
    case "/":
      return new Response(remoteIP, {
        headers: {"Content-Type": "text/plain"},
      });
      break;
    case "/json":
      return new Response(JSON.stringify({"ip": remoteIP}), {
        headers: {"Content-Type": "application/json"},
      });
      break;
    default:
      return new Response("404 Not Found", {
        status: 404,
        headers: {"Content-Type": "text/plain"},
      });
      break;
  }
}
