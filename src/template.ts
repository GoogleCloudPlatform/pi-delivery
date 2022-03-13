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

import { querySelectorOrThrow } from "lib/dom";

initCodeViewer();
initNav();

function initCodeViewer() {
  document.querySelectorAll("#sidebar a").forEach((link) => {
    link.addEventListener("click", () =>
      document.querySelector("body")?.classList.remove("nav-open")
    );
  });

  for (const codeViewer of document.querySelectorAll(".code-viewer")) {
    const codeBlocks = codeViewer.querySelectorAll<HTMLElement>(
      "div.highlighter-rouge"
    );

    querySelectorOrThrow(
      codeViewer,
      HTMLElement,
      "div.highlighter-rouge:first-child"
    ).style.display = "block";
    codeViewer
      .querySelectorAll<HTMLElement>("div.highlighter-rouge:not(:first-child)")
      .forEach((e) => (e.style.display = "none"));

    const ul = document.createElement("ul");
    ul.classList.add("languages");

    codeBlocks.forEach((block) => {
      const li = document.createElement("li");
      const a = document.createElement("a");
      a.textContent = block.title;
      a.addEventListener("click", () => {
        ul.querySelectorAll("a").forEach((e) => e.classList.remove("active"));
        a.classList.add("active");
        codeBlocks.forEach((e) => (e.style.display = "none"));
        block.style.display = "block";
      });
      li.appendChild(a);
      ul.appendChild(li);
    });
    ul.firstElementChild?.firstElementChild?.classList.add("active");
    codeViewer.prepend(ul);
  }
}

function initNav() {
  const navLinks = [
    ...document.querySelectorAll<HTMLAnchorElement>(".sidebar a.nav-item"),
  ].reduce(
    (prev, curr) => prev.set(curr.dataset.docId ?? "", curr),
    new Map<string, HTMLAnchorElement>()
  );
  const articles = [
    ...document.querySelectorAll<HTMLElement>(".main .doc-content"),
  ];

  if (articles.length === 0) return;

  let visible = [articles[0]];

  const observerCallback = (entries: IntersectionObserverEntry[]) => {
    const set = new Set(visible);
    for (const e of entries) {
      if (e.isIntersecting) {
        set.add(e.target as HTMLElement);
      } else {
        set.delete(e.target as HTMLElement);
      }
    }
    visible = [...set].sort(
      (a, b) => a.getBoundingClientRect().top - b.getBoundingClientRect().top
    );
  };

  let ticking = false;
  const scrollListener = () => {
    if (!ticking) {
      window.requestAnimationFrame(() => {
        observerCallback(observer.takeRecords());
        navLinks.forEach((l) => l.classList.remove("active"));
        let active = visible[0];
        for (const section of visible) {
          if (
            section.getBoundingClientRect().top / (window.innerHeight || 1) <
            0.2
          ) {
            active = section;
          }
        }
        navLinks.get(active.dataset.docId ?? "")?.classList.add("active");
        ticking = false;
      });
      ticking = true;
    }
  };

  const hashChangeListener = () => {
    const docId = location.hash.substring(1);
    const link = navLinks.get(docId);
    if (link) {
      navLinks.forEach((l) => l.classList.remove("active"));
      link.classList.add("active");
      const section = articles.find((v) => v.dataset.docId === docId);
      if (section && !visible.includes(section)) {
        visible = [section, ...visible];
      }
    } else {
      document.dispatchEvent(new Event("scroll"));
    }
  };

  navLinks.forEach((e) => {
    e.addEventListener("click", (ev) => {
      ev.preventDefault();
      const docId = e.dataset.docId;
      document
        .querySelector<HTMLElement>(`.main #${docId}`)
        ?.scrollIntoView({ behavior: "smooth" });
      if (docId) {
        const state = { activeDoc: docId };
        history.pushState(state, "", `#${docId}`);
      }
    });
  });

  const observer = new IntersectionObserver(observerCallback);

  articles.forEach((article) => {
    observer.observe(article);
  });

  document.addEventListener("scroll", scrollListener);
  window.addEventListener("hashchange", hashChangeListener);
  scrollListener();
}

export {};
