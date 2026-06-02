'use client';

import { useEffect, useRef } from 'react';

// Lorenz system parameters — classic chaos butterfly
const SIGMA = 10;
const RHO   = 28;
const BETA  = 8 / 3;
const DT    = 0.006;
const STEPS = 7;       // Lorenz steps drawn per animation frame
const SCALE = 4.8;     // world-unit scale factor

export function LorenzAttractor() {
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
      const renderer = new THREE.WebGLRenderer({ alpha: true, antialias: true });
      renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));
      renderer.setSize(W, H);
      renderer.autoClear = false;   // trail accumulation
      el.appendChild(renderer.domElement);

      const camera = new THREE.PerspectiveCamera(50, W / H, 1, 2000);
      camera.position.set(0, 0, 340);

      const scene = new THREE.Scene();
      const dark  = document.documentElement.classList.contains('dark');

      // ── Screen-space fade plane ───────────────────────────────────────────────
      const fadeMat   = new THREE.MeshBasicMaterial({
        color:       dark ? 0x0a0a0a : 0xf8f8f8,
        transparent: true,
        opacity:     0.022,     // very slow fade → long luminous trails
      });
      const fadeScene   = new THREE.Scene();
      const fadeCamera  = new THREE.OrthographicCamera(-1, 1, 1, -1, 0, 1);
      fadeScene.add(new THREE.Mesh(new THREE.PlaneGeometry(2, 2), fadeMat));

      // ── Trail line — redraws STEPS new segments every frame ───────────────────
      // The fade plane keeps the history visible as a decaying trail.
      const SEG       = STEPS + 1;         // vertices (1 extra to connect)
      const trailPos  = new Float32Array(SEG * 3);
      const trailGeo  = new THREE.BufferGeometry();
      const trailAttr = new THREE.BufferAttribute(trailPos, 3);
      trailAttr.setUsage(THREE.DynamicDrawUsage);
      trailGeo.setAttribute('position', trailAttr);
      trailGeo.setDrawRange(0, SEG);

      const trailMat = new THREE.LineBasicMaterial({
        color:       dark ? 0xfbbf24 : 0xd97706,
        transparent: true,
        opacity:     0.95,
      });
      scene.add(new THREE.Line(trailGeo, trailMat));

      // Bright dot at the current head of the trajectory
      const headBuf = new Float32Array(3);
      const headGeo = new THREE.BufferGeometry();
      headGeo.setAttribute('position', new THREE.BufferAttribute(headBuf, 3));
      const headMat = new THREE.PointsMaterial({
        color: 0xfbbf24, size: 7, sizeAttenuation: true,
        transparent: true, opacity: 1.0,
      });
      scene.add(new THREE.Points(headGeo, headMat));

      // ── Theme ─────────────────────────────────────────────────────────────────
      const themeObs = new MutationObserver(() => {
        const d = document.documentElement.classList.contains('dark');
        fadeMat.color.setHex(d ? 0x0a0a0a : 0xf8f8f8);
        trailMat.color.setHex(d ? 0xfbbf24 : 0xd97706);
      });
      themeObs.observe(document.documentElement, { attributes: true, attributeFilter: ['class'] });

      // ── Resize ────────────────────────────────────────────────────────────────
      const onResize = () => {
        const w = el.clientWidth, h = el.clientHeight;
        renderer.setSize(w, h);
        camera.aspect = w / h;
        camera.updateProjectionMatrix();
      };
      window.addEventListener('resize', onResize);

      // ── Lorenz state — start near but not at the origin ───────────────────────
      let lx = 0.1, ly = 0.1, lz = 0.1;
      let camT = 0;

      // Prime the system — spin up the equations silently so the butterfly
      // is already established when the user sees it.
      for (let i = 0; i < 3000; i++) {
        const dx = SIGMA * (ly - lx) * DT;
        const dy = (lx * (RHO - lz) - ly) * DT;
        const dz = (lx * ly - BETA * lz) * DT;
        lx += dx; ly += dy; lz += dz;
      }

      // Seed the first trail point
      trailPos[0] = lx * SCALE;
      trailPos[1] = (lz - 27) * SCALE;   // lz≈0–50, center at 27
      trailPos[2] = ly * SCALE;

      renderer.autoClear = true;
      renderer.clear();
      renderer.autoClear = false;

      // ── Tick ──────────────────────────────────────────────────────────────────
      const tick = () => {
        raf = requestAnimationFrame(tick);
        camT += 0.003;

        // 1. Fade — slow decay gives luminous, ghostly trails
        renderer.render(fadeScene, fadeCamera);

        // 2. Step the Lorenz equations STEPS times; record a short new segment
        for (let s = 0; s < SEG; s++) {
          const dx = SIGMA * (ly - lx) * DT;
          const dy = (lx * (RHO - lz) - ly) * DT;
          const dz = (lx * ly - BETA * lz) * DT;
          lx += dx; ly += dy; lz += dz;
          trailPos[s * 3]     = lx * SCALE;
          trailPos[s * 3 + 1] = (lz - 27) * SCALE;
          trailPos[s * 3 + 2] = ly * SCALE;
        }
        trailAttr.needsUpdate = true;

        // Update head dot
        headBuf[0] = trailPos[(SEG-1)*3];
        headBuf[1] = trailPos[(SEG-1)*3+1];
        headBuf[2] = trailPos[(SEG-1)*3+2];
        (headGeo.getAttribute('position') as import('three').BufferAttribute).needsUpdate = true;

        // 3. Slowly orbit the camera for a cinematic 3-D reveal
        camera.position.x = Math.cos(camT * 0.07) * 340;
        camera.position.z = Math.sin(camT * 0.07) * 340;
        camera.position.y = Math.sin(camT * 0.035) * 90;
        camera.lookAt(0, 0, 0);

        renderer.render(scene, camera);
      };
      tick();

      doCleanup = () => {
        cancelAnimationFrame(raf);
        themeObs.disconnect();
        window.removeEventListener('resize', onResize);
        if (el.contains(renderer.domElement)) el.removeChild(renderer.domElement);
        renderer.dispose();
        trailGeo.dispose(); trailMat.dispose();
        headGeo.dispose();  headMat.dispose();
        fadeMat.dispose();
      };
    });

    return () => { disposed = true; doCleanup?.(); };
  }, []);

  return (
    <div ref={mountRef} className="absolute inset-0 pointer-events-none" aria-hidden="true" />
  );
}
