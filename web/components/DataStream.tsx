'use client';

import { useEffect, useRef } from 'react';

const COLS        = 20;
const PTS_PER_COL = 8;

export function DataStream() {
  const mountRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const el = mountRef.current;
    if (!el) return;

    let raf = 0;
    let disposed = false;
    let doCleanup: (() => void) | undefined;

    import('three').then((THREE) => {
      if (disposed) return;

      const W = el.clientWidth;
      const H = el.clientHeight;
      const N = COLS * PTS_PER_COL;

      // ── Renderer ────────────────────────────────────────────────────────────
      const renderer = new THREE.WebGLRenderer({ alpha: true, antialias: false });
      renderer.setPixelRatio(Math.min(window.devicePixelRatio, 1.5));
      renderer.setSize(W, H);
      renderer.setClearColor(0x000000, 0);
      el.appendChild(renderer.domElement);

      // Orthographic so 1 unit = 1 pixel — easy to reason about particle positions
      const camera = new THREE.OrthographicCamera(-W/2, W/2, H/2, -H/2, 0.1, 100);
      camera.position.z = 10;

      const scene = new THREE.Scene();
      const dark  = document.documentElement.classList.contains('dark');

      // ── Particle data ───────────────────────────────────────────────────────
      const positions = new Float32Array(N * 3);
      const speeds    = new Float32Array(N);
      const homeX     = new Float32Array(N);

      const colW = W / COLS;
      for (let c = 0; c < COLS; c++) {
        const cx = -W/2 + colW * c + colW * 0.5 + (Math.random() - 0.5) * colW * 0.55;
        for (let p = 0; p < PTS_PER_COL; p++) {
          const i         = c * PTS_PER_COL + p;
          homeX[i]        = cx;
          positions[i*3]  = cx + (Math.random() - 0.5) * 16;
          positions[i*3+1]= (Math.random() - 0.5) * H;   // scattered Y at start
          positions[i*3+2]= 0;
          speeds[i] = 0.28 + Math.random() * 0.42;
        }
      }

      const geo     = new THREE.BufferGeometry();
      const posAttr = new THREE.BufferAttribute(positions, 3);
      posAttr.setUsage(THREE.DynamicDrawUsage);
      geo.setAttribute('position', posAttr);

      const mat = new THREE.PointsMaterial({
        color:           dark ? 0xfbbf24 : 0xb45309,
        size:            3,
        sizeAttenuation: false,
        transparent:     true,
        opacity:         0.20,
      });
      scene.add(new THREE.Points(geo, mat));

      // ── Theme ───────────────────────────────────────────────────────────────
      const themeObs = new MutationObserver(() => {
        mat.color.setHex(
          document.documentElement.classList.contains('dark') ? 0xfbbf24 : 0xb45309
        );
      });
      themeObs.observe(document.documentElement, { attributes: true, attributeFilter: ['class'] });

      // ── Resize ──────────────────────────────────────────────────────────────
      const onResize = () => {
        const w = el.clientWidth, h = el.clientHeight;
        renderer.setSize(w, h);
        const cam = camera as import('three').OrthographicCamera;
        cam.left   = -w/2; cam.right  =  w/2;
        cam.top    =  h/2; cam.bottom = -h/2;
        cam.updateProjectionMatrix();
      };
      window.addEventListener('resize', onResize);

      // ── Tick — particles drift upward, wrap from top → bottom ───────────────
      const halfH = H / 2;
      const tick  = () => {
        raf = requestAnimationFrame(tick);
        for (let i = 0; i < N; i++) {
          positions[i*3+1] += speeds[i];
          if (positions[i*3+1] > halfH + 8) {
            positions[i*3+1] = -halfH - Math.random() * 80;
            positions[i*3]   = homeX[i] + (Math.random() - 0.5) * 16;
          }
        }
        posAttr.needsUpdate = true;
        renderer.render(scene, camera);
      };
      tick();

      doCleanup = () => {
        cancelAnimationFrame(raf);
        themeObs.disconnect();
        window.removeEventListener('resize', onResize);
        if (el.contains(renderer.domElement)) el.removeChild(renderer.domElement);
        renderer.dispose();
        geo.dispose();
        mat.dispose();
      };
    });

    return () => { disposed = true; doCleanup?.(); };
  }, []);

  return (
    <div ref={mountRef} className="absolute inset-0 pointer-events-none" aria-hidden="true" />
  );
}
