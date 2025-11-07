'use client';

import React, { useCallback, useEffect, useMemo, useRef } from 'react';

interface AutoFitTextProps {
  value: string | number;
  maxFontSize?: number; // px
  minFontSize?: number; // px
  step?: number; // px decrement step
  fontWeight?: React.CSSProperties['fontWeight'];
  color?: string;
  lineHeight?: React.CSSProperties['lineHeight'];
  title?: string;
  style?: React.CSSProperties;
}

/**
 * AutoFitText: automatically scales down font-size so text always fits its container width.
 * - Uses ResizeObserver to re-fit on container resize
 * - Clamps between minFontSize and maxFontSize
 */
const AutoFitText: React.FC<AutoFitTextProps> = ({
  value,
  maxFontSize = 28,
  minFontSize = 14,
  step = 1,
  fontWeight = 700,
  color,
  lineHeight = 1.2,
  title,
  style,
}) => {
  const containerRef = useRef<HTMLDivElement | null>(null);
  const textRef = useRef<HTMLSpanElement | null>(null);

  const text = useMemo(() => (value ?? '').toString(), [value]);

  const fit = useCallback(() => {
    const container = containerRef.current;
    const el = textRef.current;
    if (!container || !el) return;

    // Reset to max before measuring
    let size = maxFontSize;
    el.style.fontSize = `${size}px`;

    // Guard: if container has zero width, skip
    const maxLoops = 100; // safety
    let loops = 0;
    while (el.scrollWidth > container.clientWidth && size > minFontSize && loops < maxLoops) {
      size = size - step;
      el.style.fontSize = `${size}px`;
      loops++;
    }
  }, [maxFontSize, minFontSize, step, text]);

  useEffect(() => {
    fit();
  }, [fit]);

  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;

    const ro = new ResizeObserver(() => fit());
    ro.observe(container);

    // Also refit on window font-size/zoom changes via resize
    const onResize = () => fit();
    window.addEventListener('resize', onResize);

    return () => {
      ro.disconnect();
      window.removeEventListener('resize', onResize);
    };
  }, [fit]);

  return (
    <div ref={containerRef} style={{ width: '100%', ...style }}>
      <span
        ref={textRef}
        style={{
          display: 'inline-block',
          whiteSpace: 'nowrap',
          fontWeight,
          color,
          lineHeight,
          verticalAlign: 'middle',
        }}
        title={title || text}
      >
        {text}
      </span>
    </div>
  );
};

export default AutoFitText;
