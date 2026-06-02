'use client';

import { useEffect, useRef } from 'react';

const N     = 4000;
const MAX_R = 380;
const ARMS  = 2;

export function ParticleGalaxy() {
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

      // ── Renderer ──────────────────────────────────────────────────────────────
      const renderer = new THREE.WebGLRenderer({ alpha: true, antialias: false });
      renderer.setPixelRatio(Math.min(window.devicePixelRatio, 1.5));
      renderer.setSize(W, H);
      renderer.setClearColor(0x000000, 0);
      el.appendChild(renderer.domElement);

      // Camera above the galaxy disk looking down at an angle — makes the 3D
      // depth immediately legible. The slight orbit reveals the full disk shape.
      const camera = new THREE.PerspectiveCamera(55, W / H, 1, 3000);
      camera.position.set(0, 270, 510);
      camera.lookAt(0, 0, 0);

      const scene = new THREE.Scene();
      const dark  = document.documentElement.classList.contains('dark');

      // ── Particle data ─────────────────────────────────────────────────────────
      const positions = new Float32Array(N * 3);
      const colors    = new Float32Array(N * 3);
      const radii     = new Float32Array(N);
      const angles    = new Float32Array(N);
      const speeds    = new Float32Array(N);
      const heights   = new Float32Array(N);

      for (let i = 0; i < N; i++) {
        // sqrt bias → more particles near centre (dense core)
        const r  = Math.pow(Math.random(), 0.55) * MAX_R;
        radii[i] = r;

        // Distribute across ARMS spiral arms
        const arm    = i % ARMS;
        const spiral = (r / MAX_R) * Math.PI * 3.8;
        const aOff   = (arm / ARMS) * Math.PI * 2;
        const scatter = (Math.random() - 0.5) * (0.6 + r / MAX_R) * 1.6;
        angles[i] = spiral + aOff + scatter;

        // Disk height — central bulge is taller
        heights[i] = (Math.random() - 0.5) * Math.max(55 - r * 0.13, 8);

        positions[i*3]   = Math.cos(angles[i]) * r;
        positions[i*3+1] = heights[i];
        positions[i*3+2] = Math.sin(angles[i]) * r;

        // Keplerian-ish angular speed: faster near centre
        speeds[i] = 0.00062 / Math.sqrt(Math.max(r, 25) / 70 + 0.25);

        // Vertex colours — bright amber core fading to dim amber / blue-white at edge
        const t  = r / MAX_R;
        const br = dark ? 0.45 + (1 - t) * 0.55 : (1 - t) * 0.68 + 0.15;
        colors[i*3]   = dark ? br        : br * 0.90;
        colors[i*3+1] = dark ? br * 0.78 : br * 0.50;
        colors[i*3+2] = dark ? br * 0.25 : 0;
      }

      const geo     = new THREE.BufferGeometry();
      const posAttr = new THREE.BufferAttribute(positions, 3);
      posAttr.setUsage(THREE.DynamicDrawUsage);
      geo.setAttribute('position', posAttr);
      const colAttr = new THREE.BufferAttribute(colors, 3);
      geo.setAttribute('color', colAttr);

      const mat = new THREE.PointsMaterial({
        size:            2.8,
        sizeAttenuation: true,   // near particles bigger — sells the 3-D depth
        vertexColors:    true,
        transparent:     true,
        opacity:         0.85,
      });
      scene.add(new THREE.Points(geo, mat));

      // ── Theme ─────────────────────────────────────────────────────────────────
      const themeObs = new MutationObserver(() => {
        const d = document.documentElement.classList.contains('dark');
        for (let i = 0; i < N; i++) {
          const t  = radii[i] / MAX_R;
          const br = d ? 0.45 + (1-t)*0.55 : (1-t)*0.68+0.15;
          colAttr.setXYZ(i,
            d ? br       : br * 0.90,
            d ? br*0.78  : br * 0.50,
            d ? br*0.25  : 0
          );
        }
        colAttr.needsUpdate = true;
      });
      themeObs.observe(document.documentElement, { attributes: true, attributeFilter: ['class'] });

      // ── Mouse ─────────────────────────────────────────────────────────────────
      let mx = 0, my = 0;
      const onMouse = (e: MouseEvent) => {
        mx = (e.clientX / window.innerWidth  - 0.5) * 2;
        my = (e.clientY / window.innerHeight - 0.5) * 2;
      };
      window.addEventListener('mousemove', onMouse);

      // ── Resize ────────────────────────────────────────────────────────────────
      const onResize = () => {
        const w = el.clientWidth, h = el.clientHeight;
        renderer.setSize(w, h);
        camera.aspect = w / h;
        camera.updateProjectionMatrix();
      };
      window.addEventListener('resize', onResize);

      // ── Tick ─────────────────────────────────────────────────────────────────
      let t = 0;
      const tick = () => {
        raf = requestAnimationFrame(tick);
        t += 0.003;

        for (let i = 0; i < N; i++) {
          angles[i] += speeds[i];
          const r    = radii[i];
          positions[i*3]   = Math.cos(angles[i]) * r;
          positions[i*3+2] = Math.sin(angles[i]) * r;
          // Y (height) stays constant — particles stay in their disk layer
        }
        posAttr.needsUpdate = true;

        // Camera drifts slowly + follows mouse — small tilt changes reveal depth
        camera.position.x += (mx * 55 - camera.position.x) * 0.022;
        camera.position.y +=
          (255 + Math.sin(t * 0.042) * 45 - my * 35 - camera.position.y) * 0.022;
        camera.position.z +=
          (500 + Math.cos(t * 0.055) * 65 - camera.position.z) * 0.022;
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
