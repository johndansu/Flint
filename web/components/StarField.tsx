'use client';

import { useEffect, useRef } from 'react';

// Extremely subtle 3D starfield — designed to add depth without
// competing with text. Very dim, very slow, no structure.

const N     = 220;
const DEPTH = 800;

export function StarField() {
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

      const renderer = new THREE.WebGLRenderer({ alpha: true, antialias: false });
      renderer.setPixelRatio(Math.min(window.devicePixelRatio, 1));
      renderer.setSize(W, H);
      renderer.setClearColor(0x000000, 0);
      el.appendChild(renderer.domElement);

      const camera = new THREE.PerspectiveCamera(70, W / H, 1, DEPTH * 2);
      camera.position.set(0, 0, 0);

      const scene = new THREE.Scene();
      const dark  = document.documentElement.classList.contains('dark');

      const positions = new Float32Array(N * 3);
      const colors    = new Float32Array(N * 3);

      const fillStar = (i: number, d: boolean) => {
        // depth-based brightness — far stars dimmer
        const z   = positions[i * 3 + 2];
        const t   = 1 - (-z / DEPTH);           // 1 near, 0 far
        const br  = d ? t * 0.22 : t * 0.14;
        colors[i * 3]     = d ? br : br * 0.88;
        colors[i * 3 + 1] = d ? br * 0.68 : br * 0.50;
        colors[i * 3 + 2] = d ? br * 0.08 : 0;
      };

      for (let i = 0; i < N; i++) {
        positions[i * 3]     = (Math.random() - 0.5) * 900;
        positions[i * 3 + 1] = (Math.random() - 0.5) * 600;
        positions[i * 3 + 2] = -(Math.random() * DEPTH);
        fillStar(i, dark);
      }

      const geo     = new THREE.BufferGeometry();
      const posAttr = new THREE.BufferAttribute(positions, 3);
      const colAttr = new THREE.BufferAttribute(colors, 3);
      posAttr.setUsage(THREE.DynamicDrawUsage);
      geo.setAttribute('position', posAttr);
      geo.setAttribute('color',    colAttr);

      const mat = new THREE.PointsMaterial({
        size: 1.4, sizeAttenuation: true,
        vertexColors: true, transparent: true, opacity: 0.45,
      });
      scene.add(new THREE.Points(geo, mat));

      // Theme
      const themeObs = new MutationObserver(() => {
        const d = document.documentElement.classList.contains('dark');
        for (let i = 0; i < N; i++) fillStar(i, d);
        colAttr.needsUpdate = true;
      });
      themeObs.observe(document.documentElement, { attributes: true, attributeFilter: ['class'] });

      // Resize
      const onResize = () => {
        const w = el.clientWidth, h = el.clientHeight;
        renderer.setSize(w, h);
        camera.aspect = w / h;
        camera.updateProjectionMatrix();
      };
      window.addEventListener('resize', onResize);

      // Tick — drift stars slowly toward camera, wrap at near plane
      const tick = () => {
        raf = requestAnimationFrame(tick);

        for (let i = 0; i < N; i++) {
          positions[i * 3 + 2] += 0.18;
          if (positions[i * 3 + 2] > 0) {
            positions[i * 3]     = (Math.random() - 0.5) * 900;
            positions[i * 3 + 1] = (Math.random() - 0.5) * 600;
            positions[i * 3 + 2] = -DEPTH;
          }
          fillStar(i, document.documentElement.classList.contains('dark'));
        }
        posAttr.needsUpdate = true;
        colAttr.needsUpdate = true;

        renderer.render(scene, camera);
      };
      tick();

      doCleanup = () => {
        cancelAnimationFrame(raf);
        themeObs.disconnect();
        window.removeEventListener('resize', onResize);
        if (el.contains(renderer.domElement)) el.removeChild(renderer.domElement);
        renderer.dispose();
        geo.dispose(); mat.dispose();
      };
    });

    return () => { disposed = true; doCleanup?.(); };
  }, []);

  return (
    <div ref={mountRef} className="absolute inset-0 pointer-events-none" aria-hidden="true" />
  );
}
