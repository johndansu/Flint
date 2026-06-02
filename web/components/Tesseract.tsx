'use client';

import { useEffect, useRef } from 'react';

// 4D hypercube (tesseract) projected into 3D.
// Multiple simultaneous 4D rotations make inner cubes become outer cubes —
// a topologically impossible morph that can't exist in 3D space.

const SCALE = 118;
const D4    = 3.0;   // 4D perspective distance

const VERTS: [number, number, number, number][] = [];
for (let i = 0; i < 16; i++) {
  VERTS.push([i & 1 ? 1 : -1, i & 2 ? 1 : -1, i & 4 ? 1 : -1, i & 8 ? 1 : -1]);
}

const EDGES: [number, number][] = [];
for (let i = 0; i < 16; i++) {
  for (let j = i + 1; j < 16; j++) {
    const d = i ^ j;
    if (d && (d & (d - 1)) === 0) EDGES.push([i, j]); // exactly one bit different
  }
}

export function Tesseract() {
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

      const renderer = new THREE.WebGLRenderer({ alpha: true, antialias: true });
      renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));
      renderer.setSize(W, H);
      renderer.setClearColor(0x000000, 0);
      el.appendChild(renderer.domElement);

      const camera = new THREE.PerspectiveCamera(44, W / H, 1, 2000);
      camera.position.set(0, 0, 490);

      const scene = new THREE.Scene();
      const dark  = document.documentElement.classList.contains('dark');
      const amber = () => document.documentElement.classList.contains('dark') ? 0xfbbf24 : 0xd97706;

      // ── Edge lines ───────────────────────────────────────────────────────────
      const edgePosArr  = new Float32Array(EDGES.length * 6);
      const edgeGeo     = new THREE.BufferGeometry();
      const edgePosAttr = new THREE.BufferAttribute(edgePosArr, 3);
      edgePosAttr.setUsage(THREE.DynamicDrawUsage);
      edgeGeo.setAttribute('position', edgePosAttr);

      const edgeMat = new THREE.LineBasicMaterial({ color: dark ? 0xfbbf24 : 0xd97706, transparent: true, opacity: 0.28 });
      scene.add(new THREE.LineSegments(edgeGeo, edgeMat));

      // ── Vertex dots ──────────────────────────────────────────────────────────
      const vertPosArr  = new Float32Array(16 * 3);
      const vertGeo     = new THREE.BufferGeometry();
      const vertPosAttr = new THREE.BufferAttribute(vertPosArr, 3);
      vertPosAttr.setUsage(THREE.DynamicDrawUsage);
      vertGeo.setAttribute('position', vertPosAttr);

      const vertMat = new THREE.PointsMaterial({ color: dark ? 0xfbbf24 : 0xd97706, size: 6.5, sizeAttenuation: true, transparent: true, opacity: 0.82 });
      scene.add(new THREE.Points(vertGeo, vertMat));

      // ── Theme ────────────────────────────────────────────────────────────────
      const themeObs = new MutationObserver(() => {
        const c = amber();
        edgeMat.color.setHex(c);
        vertMat.color.setHex(c);
      });
      themeObs.observe(document.documentElement, { attributes: true, attributeFilter: ['class'] });

      // ── Mouse ────────────────────────────────────────────────────────────────
      let mx = 0, my = 0;
      const onMouse = (e: MouseEvent) => {
        mx = (e.clientX / window.innerWidth  - 0.5) * 2;
        my = (e.clientY / window.innerHeight - 0.5) * 2;
      };
      window.addEventListener('mousemove', onMouse);

      // ── Resize ───────────────────────────────────────────────────────────────
      const onResize = () => {
        const w = el.clientWidth, h = el.clientHeight;
        renderer.setSize(w, h);
        camera.aspect = w / h;
        camera.updateProjectionMatrix();
      };
      window.addEventListener('resize', onResize);

      // ── 4D → 3D projection ───────────────────────────────────────────────────
      const project = (v: [number, number, number, number], t: number): [number, number, number] => {
        let x = v[0] * SCALE, y = v[1] * SCALE, z = v[2] * SCALE, w = v[3] * SCALE;

        // XW rotation
        let tmp = x * Math.cos(t * 0.61) - w * Math.sin(t * 0.61);
        w       = x * Math.sin(t * 0.61) + w * Math.cos(t * 0.61);  x = tmp;

        // YW rotation
        tmp = y * Math.cos(t * 0.43) - w * Math.sin(t * 0.43);
        w   = y * Math.sin(t * 0.43) + w * Math.cos(t * 0.43);  y = tmp;

        // ZW rotation (slow — reveals the structure evolving)
        tmp = z * Math.cos(t * 0.20) - w * Math.sin(t * 0.20);
        w   = z * Math.sin(t * 0.20) + w * Math.cos(t * 0.20);  z = tmp;

        // XY rotation (keeps the 3D projection interesting)
        tmp = x * Math.cos(t * 0.27) - y * Math.sin(t * 0.27);
        y   = x * Math.sin(t * 0.27) + y * Math.cos(t * 0.27);  x = tmp;

        // Perspective: 4D camera at distance D4 along w axis
        const s = D4 / (D4 - w / SCALE);
        return [x * s, y * s, z * s];
      };

      // ── Tick ─────────────────────────────────────────────────────────────────
      let t = 0;
      const tick = () => {
        raf = requestAnimationFrame(tick);
        t += 0.004;

        const proj = VERTS.map(v => project(v, t));

        proj.forEach(([px, py, pz], i) => {
          vertPosArr[i * 3] = px; vertPosArr[i * 3 + 1] = py; vertPosArr[i * 3 + 2] = pz;
        });
        vertPosAttr.needsUpdate = true;

        EDGES.forEach(([a, b], i) => {
          edgePosArr[i * 6]     = proj[a][0]; edgePosArr[i * 6 + 1] = proj[a][1]; edgePosArr[i * 6 + 2] = proj[a][2];
          edgePosArr[i * 6 + 3] = proj[b][0]; edgePosArr[i * 6 + 4] = proj[b][1]; edgePosArr[i * 6 + 5] = proj[b][2];
        });
        edgePosAttr.needsUpdate = true;

        camera.position.x += (mx * 58 - camera.position.x) * 0.032;
        camera.position.y += (-my * 38 - camera.position.y) * 0.032;
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
        edgeGeo.dispose(); edgeMat.dispose();
        vertGeo.dispose(); vertMat.dispose();
      };
    });

    return () => { disposed = true; doCleanup?.(); };
  }, []);

  return (
    <div ref={mountRef} className="absolute inset-0 pointer-events-none" aria-hidden="true" />
  );
}
