'use client';

import { useEffect, useRef, type ReactNode } from 'react';

interface Props {
  children: ReactNode;
  delay?: number;
  className?: string;
}

export function ScrollReveal({ children, delay = 0, className = '' }: Props) {
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const el = ref.current;
    if (!el) return;

    // Set invisible state after hydration (avoids SSR flash, preserves SEO content)
    el.style.opacity = '0';
    el.style.transform = 'translateY(26px)';
    el.style.transition = `opacity 0.72s cubic-bezier(.16,1,.3,1) ${delay}ms, transform 0.72s cubic-bezier(.16,1,.3,1) ${delay}ms`;

    const io = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          el.style.opacity = '1';
          el.style.transform = 'translateY(0)';
          io.disconnect();
        }
      },
      { threshold: 0.1, rootMargin: '0px 0px -36px 0px' }
    );
    io.observe(el);
    return () => io.disconnect();
  }, [delay]);

  return (
    <div ref={ref} className={className}>
      {children}
    </div>
  );
}
