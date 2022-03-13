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

export const scaleNames: [string, string][] = [
  ["c-major", "C Major"],
  ["e-major", "E Major"],
  ["g-major", "G Major"],
  ["cm-pentatonic", "C Minor Pentatonic"],
  ["am-pentatonic", "A Minor Pentatonic"],
];

export const scales = new Map<string, Array<string>>([
  ["c-major", ["C4", "D4", "E4", "F4", "G4", "A4", "B4", "C5", "D5", "E5"]],
  ["d-major", ["D4", "E4", "F#4", "G4", "A4", "B4", "C#5", "D5", "E5", "F#5"]],
  [
    "e-major",
    ["E4", "F#4", "G#4", "A4", "B4", "C#5", "D#5", "E5", "F#5", "G#5"],
  ],
  ["g-major", ["G4", "A4", "B4", "C5", "D5", "E5", "F#5", "G5", "A5", "B5"]],
  [
    "nine-tone",
    ["C4", "D4", "D#4", "E4", "F#4", "G4", "G#4", "A4", "B4", "C5"],
  ],
  [
    "cm-pentatonic",
    ["C4", "D#4", "F4", "G4", "A#4", "C5", "D#5", "F5", "G5", "A#5"],
  ],
  [
    "am-pentatonic",
    ["A3", "C4", "D4", "E4", "G4", "A5", "C5", "D5", "E5", "G5"],
  ],
]);

export const instruments: [number, string][] = [
  [0, "Piano"],
  [71, "Clarinet Piano Duet"],
  [73, "Flute Piano Duet"],
  [40, "Violin Piano Duet"],
];

export function bpmToDelayMs(bpm: number): number {
  return (60 / bpm) * 1000;
}

export const lowestMidiNote = 21;
export const highestMidiNote = 108;
