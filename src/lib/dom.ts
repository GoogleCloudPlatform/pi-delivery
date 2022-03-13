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

type Constructor<T> = new (...args: unknown[]) => T;

export function querySelectorOrThrow<T extends Element>(
  parent: Document | Element,
  type: Constructor<T>,
  s: string
): T {
  const elem = parent.querySelector<T>(s);
  if (!elem) {
    throw new Error(`querySelector(${s}) returned null`);
  }
  if (!(elem instanceof type)) {
    throw new Error(
      `type didn't match: expected ${typeof type}, actual ${typeof elem}`
    );
  }
  return elem;
}
