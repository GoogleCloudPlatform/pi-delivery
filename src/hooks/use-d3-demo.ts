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

import { useCallback, useEffect, useMemo, useState } from "react";
import * as d3 from "d3";

interface Props {
  ref: React.RefObject<HTMLElement>;
  width: number;
  height?: number;
  minDigit?: number;
  maxDigit?: number;
}

interface ArcData {
  value: string;
  startAngle: number;
  endAngle: number;
}

function getGradId(fromDigit: number, toDigit: number) {
  return "line-grad-" + String(fromDigit) + "-" + String(toDigit);
}

export default function useD3Demo({
  ref,
  width,
  height = width,
  minDigit = 0,
  maxDigit = 9,
}: Readonly<Props>) {
  const digitColor = useMemo(
    () =>
      d3
        .scaleOrdinal(d3.schemeTableau10)
        .domain(
          d3.range(minDigit, maxDigit).map((v: number): string => String(v))
        ),
    [maxDigit, minDigit]
  );
  const [linkNum, setLinkNum] = useState<number[]>(
    Array.from({ length: maxDigit - minDigit + 1 }, () => 0)
  );
  const [prevDigit, setPrevDigit] = useState<number>();
  const [innerRadius, setInnerRadius] = useState(width * 0.4);
  const [currentWidth, setCurrentWidth] = useState(width);
  const [currentHeight, setCurrentHeight] = useState(height);

  useEffect(() => {
    if (!ref.current) return;

    ref.current.innerHTML = "";

    const svg = d3
      .select(ref.current)
      .append("svg")
      .attr("preserveAspectRatio", "xMidYMid meet")
      .attr(
        "viewBox",
        `-${currentWidth / 2} -${
          currentHeight / 2
        } ${currentWidth} ${currentHeight}`
      );
    const g = svg.append("g");

    const arcData: ArcData[] = [];
    for (let i = minDigit; i <= maxDigit; i++) {
      arcData.push({
        value: String(i),
        startAngle: Math.PI * 0.2 * i,
        endAngle: Math.PI * 0.2 * (i + 1),
      });
    }

    g.append("defs").attr("id", "transition-defs");

    const outerRadius = innerRadius * 1.1;

    const digitArc = g
      .selectAll(".digitArc")
      .data(arcData)
      .enter()
      .append("g")
      .attr("class", "digitArc");

    const arc = d3
      .arc<ArcData>()
      .innerRadius(innerRadius)
      .outerRadius(outerRadius);

    digitArc
      .append("path")
      .attr("id", (_, i) => "digitArc" + i)
      .attr("d", arc)
      .style("fill", (_, i): string => digitColor(String(i)))
      .style("stroke", (_, i): string => digitColor(String(i)));

    const fontSize = (outerRadius - innerRadius) * 0.8;
    const digitText = digitArc
      .append("text")
      .attr("x", 4)
      .attr("dy", fontSize)
      .attr("font-size", fontSize);

    digitText
      .append("textPath")
      .attr("xlink:href", (d, i) => "#digitArc" + i)
      .text((d) => d.value);

    g.append("g").attr("class", "transition").attr("fill", "none");
  }, [
    currentHeight,
    currentWidth,
    digitColor,
    innerRadius,
    maxDigit,
    minDigit,
    ref,
  ]);

  const init = useCallback(() => {
    setLinkNum(Array.from({ length: maxDigit - minDigit + 1 }, () => 0));
    setPrevDigit(undefined);
    setCurrentWidth(width);
    setCurrentHeight(height);
    setInnerRadius(width * 0.4);
    d3.select(ref.current).selectAll("svg g.transition path").remove();
  }, [height, maxDigit, minDigit, ref, width]);

  const drawLine = useCallback(
    (fromDigit: number, toDigit: number) => {
      const svg = d3.select(ref.current);

      const line = d3
        .line()
        .x((d) => d[0])
        .y((d) => d[1])
        .curve(d3.curveBasis);

      const fromAngle = Math.PI * 0.2 * fromDigit + linkNum[fromDigit] * 0.0025;

      // If the angle has rolled over into the angle for the next digit
      // we need to reset the angle so we stay within the angle range for
      // fromDigit. Here we reset the linkNum so that we start over from the
      // original angle.
      if (fromAngle >= Math.PI * 0.2 * (fromDigit + 1)) {
        setLinkNum((l) => {
          l[fromDigit] = 0;
          return l;
        });
      }

      const toAngle = Math.PI * 0.2 * toDigit + linkNum[toDigit] * 0.0025;
      if (toAngle >= Math.PI * 0.2 * (toDigit + 1)) {
        setLinkNum((l) => {
          l[toDigit] = 0;
          return l;
        });
      }

      const fromPoint: [number, number] = [
        innerRadius * Math.sin(fromAngle),
        innerRadius * Math.cos(fromAngle) * -1,
      ];

      const middlePoint: [number, number] = [
        // TODO: get true halfway point between to radian values
        0, 0,
      ];

      const toPoint: [number, number] = [
        innerRadius * Math.sin(toAngle),
        innerRadius * Math.cos(toAngle) * -1,
      ];

      const gradId = getGradId(fromDigit, toDigit);

      if (!svg.select(gradId).node()) {
        svg
          .select("#transition-defs")
          .append("linearGradient")
          .attr("id", gradId)
          .attr("gradientUnits", "userSpaceOnUse")
          .attr("x1", fromPoint[0])
          .attr("y1", fromPoint[1])
          .attr("x2", toPoint[0])
          .attr("y2", toPoint[1])
          .selectAll("stop")
          .data([fromDigit, toDigit])
          .enter()
          .append("stop")
          .attr("offset", (_, i) => i * 100 + "%")
          .attr("stop-color", (d) => d3.rgb(digitColor(String(d))).formatHex());
      }

      // finally add the line path element
      const transitions = svg.select(".transition");
      transitions
        .append("path")
        .attr("d", line([fromPoint, middlePoint, toPoint]))
        .attr("stroke", () => "url(#" + getGradId(fromDigit, toDigit) + ")")
        .style("stroke-opacity", 0.75);

      setLinkNum((l) => {
        l[fromDigit]++;
        l[toDigit]++;
        return l;
      });
    },
    [digitColor, innerRadius, linkNum, ref]
  );

  const draw = useCallback(
    (d: number) => {
      if (prevDigit) {
        drawLine(prevDigit, d);
      }
      setPrevDigit(d);
    },
    [prevDigit, drawLine]
  );

  /*
  useEffect(() => {
    const svg = d3.select(ref.current).select("svg");
    svg.attr("width", width).attr("height", height);
    svg.select("g").attr("transform", `translate(${width / 2}, ${height / 2})`);
  }, [width, height, ref]);
*/
  return { init, draw };
}
