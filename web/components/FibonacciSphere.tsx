'use client';

import { useEffect, useRef } from 'react';

// 700 particles distributed on a sphere using the golden-angle (Fibonacci)
// method — produces a perfectly uniform sunflower pattern.
// Dim and slow — designed for text-heavy sections.

const N      = 700;
const R      = 200;
const GOLDEN = Math.PI * (3 - Math.sqrt(5));  // ≈ 2.39996 rad (golden angle)

export function FibonacciSphere() {
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
      renderer.setPixelRatio(Math.min(window.devicePixelRatio, 1.5));
      renderer.setSize(W, H);
      renderer.setClearColor(0x000000, 0);
      el.appendChild(renderer.domElement);

      const camera = new THREE.PerspectiveCamera(55, W / H, 1, 2000);
      camera.position.set(0, 0, 440);
      camera.lookAt(0, 0, 0);

      const scene  = new THREE.Scene();
      const dark   = document.documentElement.classList.contains('dark');

      const positions = new Float32Array(N * 3);
      const colors    = new Float32Array(N * 3);

      const computeColor = (i: number, d: boolean) => {
        // latitude-based brightness: bright at top, dim at bottom
        const y   = positions[i * 3 + 1];
        const lat = y / R;                     // -1 (south) to +1 (north)
        const br  = d
          ? 0.08 + (lat * 0.5 + 0.5) * 0.28   // 0.08 → 0.36
          : 0.05 + (lat * 0.5 + 0.5) * 0.20;  // 0.05 → 0.25
        colors[i * 3]     = d ? br         : br * 0.90;
        colors[i * 3 + 1] = d ? br * 0.62  : br * 0.46;
        colors[i * 3 + 2] = d ? br * 0.06  : 0;
      };

      for (let i = 0; i < N; i++) {
        const y     = 1 - (i / (N - 1)) * 2;          // 1 down to -1
        const r     = Math.sqrt(1 - y * y) * R;
        const theta = GOLDEN * i;
        positions[i * 3]     = Math.cos(theta) * r;
        positions[i * 3 + 1] = y * R;
        positions[i * 3 + 2] = Math.sin(theta) * r;
        computeColor(i, dark);
      }

      const geo     = new THREE.BufferGeometry();
      const colAttr = new THREE.BufferAttribute(colors, 3);
      geo.setAttribute('position', new THREE.BufferAttribute(positions, 3));
      geo.setAttribute('color',    colAttr);

      const mat = new THREE.PointsMaterial({
        size: 1.8, sizeAttenuation: true,
        vertexColors: true, transparent: true, opacity: 0.60,
      });
      const pts = new THREE.Points(geo, mat);
      scene.add(pts);

      // Theme
      const themeObs = new MutationObserver(() => {
        const d = document.documentElement.classList.contains('dark');
        for (let i = 0; i < N; i++) computeColor(i, d);
        colAttr.needsUpdate = true;
      });
      themeObs.observe(document.documentElement, { attributes: true, attributeFilter: ['class'] });

      // Mouse
      let mx = 0, my = 0;
      const onMouse = (e: MouseEvent) => {
        mx = (e.clientX / window.innerWidth  - 0.5) * 2;
        my = (e.clientY / window.innerHeight - 0.5) * 2;
      };
      window.addEventListener('mousemove', onMouse);

      // Resize
      const onResize = () => {
        const w = el.clientWidth, h = el.clientHeight;
        renderer.setSize(w, h);
        camera.aspect = w / h;
        camera.updateProjectionMatrix();
      };
      window.addEventListener('resize', onResize);

      // Tick
      let t = 0;
      const tick = () => {
        raf = requestAnimationFrame(tick);
        t += 0.003;

        pts.rotation.y  =  t * 0.12;
        pts.rotation.x  =  Math.sin(t * 0.09) * 0.10;

        camera.position.x += (mx * 45 - camera.position.x) * 0.020;
        camera.position.y += (-my * 30 - camera.position.y) * 0.020;
        camera.lookAt(0, 0, 0);

        renderer.render(scene, camera);
      };
      tick();

      doCleanup = () => {
        cancelAnimationFrame(raf);
        themeObs.disconnect();
        window.removeEventListener('mousemove', onMouse);
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
