/**
 * Copyright 2022 Google LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import React, { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import mobile from "is-mobile";
import "template";

import "@fontsource/roboto/300.css";
import "@fontsource/roboto/400.css";
import "@fontsource/roboto/500.css";
import "@fontsource/roboto/700.css";

import "app.scss";

import App from "App";
import Demo from "components/Demo";

// Lazy load bigger demos
const Fetch = React.lazy(() => import("demos/Fetch"));
const Coconet = React.lazy(() => import("demos/Coconet"));
const Genie = React.lazy(() => import("demos/Genie"));
const D3 = React.lazy(() => import("demos/D3"));

const container = document.getElementById("react-root");
if (container) {
  const root = createRoot(container);
  root.render(
    <StrictMode>
      <App />
    </StrictMode>
  );
} else {
  console.error("getElementById returned nil for react-root");
}

// These demos are available on mobile devices as well.
const demos: Array<[React.ReactNode, string]> = [
  [<Fetch />, "fetch-demo"],
  [<Genie />, "genie-demo"],
  [<D3 />, "d3-demo"],
];

demos.forEach(([demo, id]) => {
  const container = document.getElementById(id);
  if (container) {
    const root = createRoot(container);
    root.render(
      <StrictMode>
        <Demo>{demo}</Demo>
      </StrictMode>
    );
  } else {
    console.error("getElementById returned nil for ", id);
  }
});

// Coconet is too big for mobile devices.
if (!mobile()) {
  const container = document.getElementById("coconet-demo");
  if (container) {
    const root = createRoot(container);
    root.render(
      <StrictMode>
        <Demo>
          <Coconet />
        </Demo>
      </StrictMode>
    );
  } else {
    console.error("getElementById returned nil for coconet-demo");
  }
}
