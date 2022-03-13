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

import { useState, useEffect, useRef } from "react";
import type * as mm from "@magenta/music";
import * as colorString from "color-string";

type Props = {
  notes: mm.INoteSequence;
  activeNote: mm.NoteSequence.INote;
  width: number;
  height: number;
  noteColor: string;
  activeNoteColor: string;
};

function colorToRgb(s: string): string{
  const rgb = colorString.get.rgb(s);
  return rgb.slice(0, 3).join(", ");
}

export default function MagentaVisualizer(
  props: Readonly<Partial<Props>> | undefined = {}
) {
  const { notes, activeNote, width, height, noteColor, activeNoteColor } = props;

  const [visualizer, setVisualizer] = useState<mm.PianoRollSVGVisualizer>();
  const svgContainerRef = useRef<SVGSVGElement>(null);

  useEffect(() => {
    const container = svgContainerRef.current;
    if (!container) return;

    async function load() {
      if (!container || !notes) return;
      const mmv = await import("@magenta/music/esm/core/visualizer");

      const noteRGB = noteColor ? colorToRgb(noteColor) : undefined;
      const activeNoteRGB = activeNoteColor ? colorToRgb(activeNoteColor) : undefined;

      setVisualizer(
        new mmv.PianoRollSVGVisualizer(notes, container, {
          noteRGB,
          activeNoteRGB,
        })
      );
    }
    load();
  }, [activeNoteColor, noteColor, notes]);

  useEffect(() => {
    if (activeNote) {
      visualizer?.redraw(activeNote);
    } else {
      visualizer?.clearActiveNotes();
    }
  }, [visualizer, activeNote]);

  return <svg ref={svgContainerRef} width={width} height={height}></svg>;
}
