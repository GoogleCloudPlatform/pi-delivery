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

import { useEffect, useMemo, useRef, useState } from "react";
import { useResizeDetector } from "react-resize-detector";
import { querySelectorOrThrow } from "lib/dom";
import GenieButtons from "./GenieButtons";
import usePianoGenie from "hooks/use-piano-genie";
import useMagentaPlayer from "hooks/use-magenta-player";

// Taken from the original Genie demo
// https://glitch.com/edit/#!/piano-genie

const CONSTANTS = {
  COLORS: [
    "#EE2B29",
    "#ff9800",
    "#ffff00",
    "#c6ff00",
    "#00e5ff",
    "#2979ff",
    "#651fff",
    "#d500f9",
  ],
  NUM_BUTTONS: 8,
  NOTES_PER_OCTAVE: 12,
  WHITE_NOTES_PER_OCTAVE: 7,
  LOWEST_PIANO_KEY_MIDI_NOTE: 21,
};

const defaultTemperature = 0.25;

interface Note {
  x: number;
  y: number;
  width: number;
  height: number;
  color: string;
  on: boolean;
}

/*************************
 * Floaty notes
 ************************/
class FloatyNotes {
  notes = Array<Note>(); // the notes floating on the screen.
  parent: HTMLElement;
  canvas: HTMLCanvasElement;
  context: CanvasRenderingContext2D;
  contextHeight = 0;
  detached = false;
  prevTimestamp: DOMHighResTimeStamp;

  constructor(parent: HTMLElement) {
    this.parent = parent;
    this.canvas = querySelectorOrThrow(parent, HTMLCanvasElement, "canvas");
    const context = this.canvas.getContext("2d");
    if (!context) throw new Error();
    context.lineWidth = 4;
    context.lineCap = "round";
    this.context = context;
    this.prevTimestamp = 0;
  }

  resize(whiteNoteHeight) {
    const rect = this.parent.getBoundingClientRect();
    this.canvas.width = rect.width;
    this.canvas.height = this.contextHeight = rect.height - whiteNoteHeight;
  }

  addNote(button, x, width) {
    const noteToPaint = {
      x: parseFloat(x),
      y: 0,
      width: parseFloat(width),
      height: 0,
      color: CONSTANTS.COLORS[button],
      on: true,
    };
    this.notes.push(noteToPaint);
    return noteToPaint;
  }

  stopNote(noteToPaint) {
    noteToPaint.on = false;
  }

  startDrawLoop() {
    this.prevTimestamp = performance.now();
    window.requestAnimationFrame(this.drawLoop.bind(this));
  }

  drawLoop(timestamp: DOMHighResTimeStamp) {
    const dy = 4 * (timestamp - this.prevTimestamp) * (1 / 60);
    this.prevTimestamp = timestamp;
    this.context.clearRect(0, 0, this.canvas.width, this.canvas.height);

    // Remove all the notes that will be off the page;
    this.notes = this.notes.filter(
      (note) => note.on || note.y < this.contextHeight
    );

    // Advance all the notes.
    for (let i = 0; i < this.notes.length; i++) {
      const note = this.notes[i];

      // If the note is still on, then its height goes up but it
      // doesn't start sliding down yet.
      if (note.on) {
        note.height += dy;
      } else {
        note.y += dy;
      }

      this.context.globalAlpha = 1 - note.y / this.contextHeight;
      this.context.fillStyle = note.color;
      this.context.fillRect(note.x, note.y, note.width, note.height);
    }
    if (!this.detached) window.requestAnimationFrame(this.drawLoop.bind(this));
  }
}

class Piano {
  config = {
    whiteNoteWidth: 20,
    blackNoteWidth: 20,
    whiteNoteHeight: 70,
    blackNoteHeight: (2 * 70) / 3,
    octaves: 7,
  };
  svg: SVGSVGElement;
  readonly svgNS = "http://www.w3.org/2000/svg";
  parent: HTMLElement;

  constructor(parent: HTMLElement) {
    this.parent = parent;
    this.svg = querySelectorOrThrow(parent, SVGSVGElement, "svg");
  }

  resize(octaves, totalWhiteNotes) {
    // i honestly don't know why some flooring is good and some is bad sigh.
    const width = this.parent.getBoundingClientRect().width;
    const ratio = width / totalWhiteNotes;
    this.config.octaves = octaves;
    this.config.whiteNoteWidth =
      this.config.octaves > 6 ? ratio : Math.floor(ratio);
    this.config.blackNoteWidth = (this.config.whiteNoteWidth * 2) / 3;
    this.svg.setAttribute("width", String(width));
    this.svg.setAttribute("height", String(this.config.whiteNoteHeight));
  }

  draw() {
    this.svg.innerHTML = "";
    const halfABlackNote = this.config.blackNoteWidth / 2;
    let x = 0;
    const y = 0;
    let index = 0;

    const blackNoteIndexes = [1, 3, 6, 8, 10];

    // First draw all the white notes.
    // Pianos start on an A (if we're using all the octaves);
    if (this.config.octaves > 6) {
      this.makeRect(
        0,
        x,
        y,
        this.config.whiteNoteWidth,
        this.config.whiteNoteHeight,
        "white",
        "#141E30"
      );
      this.makeRect(
        2,
        this.config.whiteNoteWidth,
        y,
        this.config.whiteNoteWidth,
        this.config.whiteNoteHeight,
        "white",
        "#141E30"
      );
      index = 3;
      x = 2 * this.config.whiteNoteWidth;
    } else {
      // Starting 3 semitones up on small screens (on a C), and a whole octave up.
      index = 3 + CONSTANTS.NOTES_PER_OCTAVE;
    }

    // Draw the white notes.
    for (let o = 0; o < this.config.octaves; o++) {
      for (let i = 0; i < CONSTANTS.NOTES_PER_OCTAVE; i++) {
        if (blackNoteIndexes.indexOf(i) === -1) {
          this.makeRect(
            index,
            x,
            y,
            this.config.whiteNoteWidth,
            this.config.whiteNoteHeight,
            "white",
            "#141E30"
          );
          x += this.config.whiteNoteWidth;
        }
        index++;
      }
    }

    if (this.config.octaves > 6) {
      // And an extra C at the end (if we're using all the octaves);
      this.makeRect(
        index,
        x,
        y,
        this.config.whiteNoteWidth,
        this.config.whiteNoteHeight,
        "white",
        "#141E30"
      );

      // Now draw all the black notes, so that they sit on top.
      // Pianos start on an A:
      this.makeRect(
        1,
        this.config.whiteNoteWidth - halfABlackNote,
        y,
        this.config.blackNoteWidth,
        this.config.blackNoteHeight,
        "black"
      );
      index = 3;
      x = this.config.whiteNoteWidth;
    } else {
      // Starting 3 semitones up on small screens (on a C), and a whole octave up.
      index = 3 + CONSTANTS.NOTES_PER_OCTAVE;
      x = -this.config.whiteNoteWidth;
    }

    // Draw the black notes.
    for (let o = 0; o < this.config.octaves; o++) {
      for (let i = 0; i < CONSTANTS.NOTES_PER_OCTAVE; i++) {
        if (blackNoteIndexes.indexOf(i) !== -1) {
          this.makeRect(
            index,
            x + this.config.whiteNoteWidth - halfABlackNote,
            y,
            this.config.blackNoteWidth,
            this.config.blackNoteHeight,
            "black"
          );
        } else {
          x += this.config.whiteNoteWidth;
        }
        index++;
      }
    }
  }

  highlightNote(note, button) {
    // Show the note on the piano roll.
    const rect = this.svg.querySelector(`rect[data-index="${note}"]`);
    if (!rect) {
      console.log("couldnt find a rect for note", note);
      return;
    }
    rect.setAttribute("active", "true");
    rect.setAttribute("class", `color-${button}`);
    return rect;
  }

  clearNote(note) {
    const rect = this.svg.querySelector(`rect[data-index="${note}"]`);
    if (!rect) {
      console.log("couldnt find a rect for note", note);
      return;
    }
    rect.removeAttribute("active");
    rect.removeAttribute("class");
  }

  makeRect(index, x, y, w, h, fill, stroke?): SVGRectElement {
    const rect = document.createElementNS(this.svgNS, "rect");
    rect.setAttribute("data-index", index);
    rect.setAttribute("x", x);
    rect.setAttribute("y", y);
    rect.setAttribute("width", w);
    rect.setAttribute("height", h);
    rect.setAttribute("fill", fill);
    if (stroke) {
      rect.setAttribute("stroke", stroke);
      rect.setAttribute("stroke-width", "3px");
    }
    this.svg.appendChild(rect);
    return rect;
  }
}

function getAvailableNotes(octaves: number): Array<number> {
  const bonusNotes = octaves > 6 ? 4 : 0; // starts on an A, ends on a C.
  const totalNotes = CONSTANTS.NOTES_PER_OCTAVE * octaves + bonusNotes;
  return Array<number>(totalNotes)
    .fill(0)
    .map((x, i) => {
      if (octaves > 6) return i;
      // Starting 3 semitones up on small screens (on a C), and a whole octave up.
      return i + 3 + CONSTANTS.NOTES_PER_OCTAVE;
    });
}

interface Props {
  activeButton?: number;
  inView?: boolean;
  onReady?: () => void;
}

export default function GeniePiano(props: Props) {
  const { activeButton, inView, onReady } = props;

  const [piano, setPiano] = useState<Piano>();
  const [floatyNotes, setFloatyNotes] = useState<FloatyNotes>();
  const [activeNote, setActiveNote] = useState<number>();
  const [noteToPaint, setNoteToPaint] = useState<Note>();
  const [octaves, setOctaves] = useState(7);
  const genie = usePianoGenie(inView);
  const {
    playNoteDown,
    playNoteUp,
    ready: playerReady,
  } = useMagentaPlayer({
    load: inView,
  });
  const availableNotes = useMemo(() => getAvailableNotes(octaves), [octaves]);
  const ref = useRef<HTMLDivElement>(null);
  const { width, height } = useResizeDetector({
    targetRef: ref,
    refreshMode: "throttle",
    refreshRate: 5,
  });

  useEffect(() => {
    if (!ref.current) return;
    const floatyNotes = new FloatyNotes(ref.current);
    floatyNotes.startDrawLoop();
    setFloatyNotes(floatyNotes);
    const piano = new Piano(ref.current);
    setPiano(piano);

    return () => {
      setPiano(undefined);
      floatyNotes.detached = true;
      setFloatyNotes(undefined);
    };
  }, []);

  useEffect(() => {
    if (!width || !piano) return;
    const octaves = width > 700 ? 7 : 3;
    setOctaves(octaves);
    const bonusNotes = octaves > 6 ? 4 : 0; // starts on an A, ends on a C.
    const totalWhiteNotes =
      CONSTANTS.WHITE_NOTES_PER_OCTAVE * octaves + (bonusNotes - 1);
    piano.resize(octaves, totalWhiteNotes);
    piano.draw();
    floatyNotes?.resize(piano.config.whiteNoteHeight);
  }, [floatyNotes, piano, width]);

  useEffect(() => {
    floatyNotes?.resize(piano?.config.whiteNoteHeight);
  }, [floatyNotes, piano, height]);

  useEffect(() => {
    if (!genie || !piano || !floatyNotes) return;
    let note: number;
    let noteToPaint: Note;
    if (activeButton !== undefined) {
      if (octaves > 6) {
        note = genie.next(activeButton);
      } else {
        note = genie.nextFromKeyList(
          activeButton,
          availableNotes,
          defaultTemperature
        );
      }
      const midi = note + CONSTANTS.LOWEST_PIANO_KEY_MIDI_NOTE;
      playNoteDown({ pitch: midi });
      const rect = piano.highlightNote(note, activeButton);
      if (rect) {
        noteToPaint = floatyNotes.addNote(
          activeButton,
          rect.getAttribute("x"),
          rect.getAttribute("width")
        );
      }
    }
    setActiveNote((prev) => {
      if (prev) {
        const midi = prev + CONSTANTS.LOWEST_PIANO_KEY_MIDI_NOTE;
        playNoteUp({ pitch: midi });
        piano?.clearNote(prev);
      }
      return note;
    });
    setNoteToPaint((prev) => {
      if (prev) {
        prev.on = false;
      }
      return noteToPaint;
    });
  }, [
    activeButton,
    availableNotes,
    floatyNotes,
    genie,
    octaves,
    piano,
    playNoteDown,
    playNoteUp,
  ]);

  useEffect(() => {
    if (onReady && genie && playerReady) onReady();
  }, [genie, onReady, playerReady]);

  return (
    <>
      <div ref={ref} className="genie-piano">
        <div className="background"></div>
        <canvas></canvas>
        <svg></svg>
      </div>
      <GenieButtons active={activeButton}></GenieButtons>
    </>
  );
}
