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

const TotalDigits = {
  10: 1e14,
  16: 83_048_202_372_185,
};

export interface PiConfig {
  url: string;
  radix: number;
}

export class Pi {
  #url: string;
  radix: number;

  constructor(config: Partial<PiConfig> = {}) {
    const DEFAULT_URL = "https://api.pi.delivery/v1/pi";

    this.#url = config.url ?? DEFAULT_URL;
    this.radix = config.radix ?? 10;
  }

  async get(
    start: number,
    numberOfDigits: number,
    callback?: (content: string) => void
  ): Promise<string> {
    const response = await fetch(
      `${this.#url}?start=${start}&numberOfDigits=${numberOfDigits}&radix=${
        this.radix
      }`
    );
    const data = await response.json();
    if (callback) {
      callback(data.content);
    }
    return data.content;
  }

  get length(): number {
    return TotalDigits[this.radix];
  }
}

interface PiStreamConfig {
  streamBufferSize: number;
  streamChunkSize: number;
  delayMs: number;
  start: number;
}

interface Chunk {
  start: number;
  digits: string;
}

export interface DigitEventDetail {
  position: number;
  digit: number;
}

type StreamEventType = "digit";
export type DigitEventHandler = (event: CustomEvent<DigitEventDetail>) => void;

export class PiStream {
  #streaming = false;
  #pi: Pi;
  #off: number;
  #bufBase: number;
  #buffer: Array<Chunk>;
  #eventTarget = new EventTarget();
  #interval: number | null = null;
  readonly config: Readonly<PiStreamConfig>;
  #delayMs: number;

  constructor(pi?: Pi, config: Partial<PiStreamConfig> = {}) {
    const DEFAULT_STREAM_BUFFER_SIZE = 2;
    const DEFAULT_STREAM_CHUNK_SIZE = 100;
    const DEFAULT_DELAY_MS = 1000;
    const DEFAULT_START = 0;

    this.config = {
      streamBufferSize: config.streamBufferSize ?? DEFAULT_STREAM_BUFFER_SIZE,
      streamChunkSize: config.streamChunkSize ?? DEFAULT_STREAM_CHUNK_SIZE,
      delayMs: config.delayMs ?? DEFAULT_DELAY_MS,
      start: config.start ?? DEFAULT_START,
    };

    this.#pi = pi ?? new Pi();
    this.#off = this.config.start;
    this.#bufBase = this.#off;
    this.#buffer = new Array<Chunk>(this.config.streamBufferSize);
    this.#delayMs = this.config.delayMs;
  }

  // #fire sends
  #fire(position: number, digit: number): void {
    const e = new CustomEvent<DigitEventDetail>("digit", {
      detail: { position: position, digit: digit },
    });
    this.#eventTarget.dispatchEvent(e);
  }

  // #fetch fetches a single chunk at digit start and stores it to streamBuffer[chunkIdx].
  // Returns streamBuffer[chunkIdx];
  async #fetch(chunkIdx: number, start: number, nChunks = 1): Promise<Chunk> {
    if (this.#buffer[chunkIdx]?.start === start) {
      return this.#buffer[chunkIdx];
    }
    if (nChunks > this.config.streamBufferSize) {
      throw new Error("nChunks bigger than buffer size");
    }
    const digits = await this.#pi.get(
      start,
      this.config.streamChunkSize * nChunks
    );
    for (let i = 0; i < nChunks; i++) {
      const pos = i * this.config.streamChunkSize;
      const chunk = {
        start: start + pos,
        digits: digits.slice(pos, pos + this.config.streamChunkSize),
      };
      this.#buffer[(chunkIdx + i) % this.config.streamBufferSize] = chunk;
    }
    return this.#buffer[chunkIdx];
  }

  #offToBuffer(off: number): {
    bufIdx: number;
    start: number;
    digitIdx: number;
  } {
    const absBufIdx = Math.floor(
      (off - this.#bufBase) / this.config.streamChunkSize
    );
    const bufIdx = absBufIdx % this.config.streamBufferSize;
    const start = this.#bufBase + absBufIdx * this.config.streamChunkSize;
    const digitIdx = off - start;
    return { bufIdx, start, digitIdx };
  }

  #startInterval() {
    const onInterval = async () => {
      const off = this.#off++;
      if (off >= this.#pi.length) {
        this.stop();
      }

      const { bufIdx, start, digitIdx } = this.#offToBuffer(off);
      const chunk = await this.#fetch(bufIdx, start);
      const digits = chunk.digits;

      const d = parseInt(digits[digitIdx]);
      if (!isNaN(d)) {
        this.#fire(off, d);
      }

      if (off >= this.#pi.length) {
        return;
      }

      if (digitIdx >= digits.length - 1) {
        this.#fetch(
          bufIdx,
          start + this.config.streamChunkSize * this.config.streamBufferSize
        );
      }
    };

    onInterval();
    this.#interval = setInterval(onInterval, this.#delayMs);
  }

  stop(): void {
    this.#streaming = false;
    if (this.#interval) {
      clearInterval(this.#interval);
      this.#interval = null;
    }
  }

  async start(): Promise<void> {
    if (this.#streaming) return;
    this.#streaming = true;

    const { bufIdx, start } = this.#offToBuffer(this.#off);
    if (this.#buffer[bufIdx]?.start !== start) {
      this.#bufBase = this.#off;
      await this.#fetch(0, this.#bufBase, this.config.streamBufferSize);
    }
    this.#startInterval();
  }

  seek(n: number): void {
    if (n !== this.#off) {
      this.#off = n;
      this.#bufBase = n;
    }
  }

  // This is for backward compatibility.
  listen(listener: DigitEventHandler): void {
    this.addEventListener("digit", listener as EventListener);
  }

  addEventListener(
    type: StreamEventType,
    listener: EventListenerOrEventListenerObject,
    options?: boolean | AddEventListenerOptions
  ) {
    this.#eventTarget.addEventListener(type, listener, options);
  }

  removeEventListener(
    type: StreamEventType,
    listener: EventListenerOrEventListenerObject,
    options?: boolean | AddEventListenerOptions
  ) {
    this.#eventTarget.removeEventListener(type, listener, options);
  }

  dispatchEvent(type: StreamEventType, eventInitDict?: EventInit): boolean {
    return this.#eventTarget.dispatchEvent(new Event(type, eventInitDict));
  }

  get delayMs(): number {
    return this.#delayMs;
  }

  set delayMs(delay: number) {
    this.#delayMs = delay;
    if (this.#interval) {
      clearInterval(this.#interval);
      this.#startInterval();
    }
  }

  get length(): number {
    return this.#pi.length;
  }
}
