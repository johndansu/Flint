'use client';

import { useEffect, useRef } from 'react';

// Smooth vector field using overlapping sine waves — no Perlin dep needed
function fieldAngle(x: number, y: number, t: number): number {
  return (
    Math.sin(x * 0.0032 + t * 0.38) * Math.cos(y * 0.0028 + t * 0.27) * Math.PI * 2.8 +
    Math.sin(x * 0.0075 + t * 0.62) * Math.PI * 0.45 +
    Math.cos(y * 0.006  + t * 0.18) * Math.PI * 0.3
  );
}

const N     = 3000;
const SPEED = 2.2;

export function FlowField() {
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

      // ── Renderer — autoClear disabled so old frames persist for trails ────────
      const renderer = new THREE.WebGLRenderer({ alpha: true, antialias: false });
      renderer.setPixelRatio(Math.min(window.devicePixelRatio, 1.5));
      renderer.setSize(W, H);
      renderer.autoClear = false;
      el.appendChild(renderer.domElement);

      // Main camera: orthographic, 1 unit = 1 pixel
      const camera = new THREE.OrthographicCamera(-W/2, W/2, H/2, -H/2, 0.1, 100);
      camera.position.z = 10;

      const scene = new THREE.Scene();

      // ── Screen-space fade plane (creates the trail effect) ───────────────────
      // Uses its own NDC-space camera so it's always full-screen regardless of resize
      const dark      = document.documentElement.classList.contains('dark');
      const fadeMat   = new THREE.MeshBasicMaterial({
        color:       dark ? 0x0c0c0c : 0xf2f3f5,
        transparent: true,
        opacity:     0.042,        // trail decay rate: lower = longer tails
      });
      const fadeMesh    = new THREE.Mesh(new THREE.PlaneGeometry(2, 2), fadeMat);
      const fadeScene   = new THREE.Scene();
      const fadeCamera  = new THREE.OrthographicCamera(-1, 1, 1, -1, 0, 1);
      fadeScene.add(fadeMesh);

      // ── Particle data ─────────────────────────────────────────────────────────
      const px       = new Float32Array(N);
      const py       = new Float32Array(N);
      const pvx      = new Float32Array(N);
      const pvy      = new Float32Array(N);
      const positions = new Float32Array(N * 3);

      const halfW = W / 2, halfH = H / 2;
      for (let i = 0; i < N; i++) {
        px[i]  = (Math.random() - 0.5) * W;
        py[i]  = (Math.random() - 0.5) * H;
        pvx[i] = 0; pvy[i] = 0;
      }

      const geo     = new THREE.BufferGeometry();
      const posAttr = new THREE.BufferAttribute(positions, 3);
      posAttr.setUsage(THREE.DynamicDrawUsage);
      geo.setAttribute('position', posAttr);

      const mat = new THREE.PointsMaterial({
        color:           dark ? 0xfbbf24 : 0xb45309,
        size:            2.2,
        sizeAttenuation: false,
        transparent:     true,
        opacity:         0.72,
      });
      scene.add(new THREE.Points(geo, mat));

      // ── Theme ─────────────────────────────────────────────────────────────────
      const themeObs = new MutationObserver(() => {
        const d = document.documentElement.classList.contains('dark');
        mat.color.setHex(d ? 0xfbbf24 : 0xb45309);
        fadeMat.color.setHex(d ? 0x0c0c0c : 0xf2f3f5);
      });
      themeObs.observe(document.documentElement, { attributes: true, attributeFilter: ['class'] });

      // ── Resize ────────────────────────────────────────────────────────────────
      const onResize = () => {
        const w = el.clientWidth, h = el.clientHeight;
        renderer.setSize(w, h);
        const cam = camera as import('three').OrthographicCamera;
        cam.left = -w/2; cam.right = w/2;
        cam.top  =  h/2; cam.bottom = -h/2;
        cam.updateProjectionMatrix();
      };
      window.addEventListener('resize', onResize);

      // ── Tick ──────────────────────────────────────────────────────────────────
      let t = 0;

      // One clean frame before accumulation starts
      renderer.autoClear = true;
      renderer.clear();
      renderer.autoClear = false;

      const tick = () => {
        raf = requestAnimationFrame(tick);
        t += 0.0038;

        // 1. Render fade plane — dims everything by ~4% → creates trails
        renderer.render(fadeScene, fadeCamera);

        // 2. Advance particles along the flow field
        for (let i = 0; i < N; i++) {
          const a    = fieldAngle(px[i], py[i], t);
          pvx[i]    += (Math.cos(a) * SPEED - pvx[i]) * 0.11;
          pvy[i]    += (Math.sin(a) * SPEED - pvy[i]) * 0.11;
          px[i]     += pvx[i];
          py[i]     += pvy[i];

          // Wrap at boundaries (seamless loop)
          if (px[i] > halfW)  px[i] -= W;
          if (px[i] < -halfW) px[i] += W;
          if (py[i] > halfH)  py[i] -= H;
          if (py[i] < -halfH) py[i] += H;

          positions[i*3]   = px[i];
          positions[i*3+1] = py[i];
          positions[i*3+2] = 0;
        }
        posAttr.needsUpdate = true;

        // 3. Render particles on top of the faded previous frame
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
        fadeMesh.geometry.dispose(); fadeMat.dispose();
      };
    });

    return () => { disposed = true; doCleanup?.(); };
  }, []);

  return (
    <div ref={mountRef} className="absolute inset-0 pointer-events-none" aria-hidden="true" />
  );
}
