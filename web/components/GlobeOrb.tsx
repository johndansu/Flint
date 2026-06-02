'use client';

import { useEffect, useRef } from 'react';

export function GlobeOrb() {
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

      // ── Renderer ────────────────────────────────────────────────────────────
      const renderer = new THREE.WebGLRenderer({ alpha: true, antialias: true });
      renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));
      renderer.setSize(W, H);
      renderer.setClearColor(0x000000, 0);
      el.appendChild(renderer.domElement);

      const camera = new THREE.PerspectiveCamera(50, W / H, 1, 2000);
      camera.position.set(0, 0, 520);

      const scene = new THREE.Scene();

      // ── Outer wireframe icosahedron ─────────────────────────────────────────
      // detail=3 → 320 faces, dense enough to look spherical
      const outerGeo = new THREE.IcosahedronGeometry(170, 3);
      const outerMat = new THREE.MeshBasicMaterial({
        color: 0xfbbf24,
        wireframe: true,
        transparent: true,
        opacity: 0.09,
      });
      const outerMesh = new THREE.Mesh(outerGeo, outerMat);
      scene.add(outerMesh);

      // ── Inner solid icosahedron (detail=1) — subtle amber glow core ─────────
      const innerGeo = new THREE.IcosahedronGeometry(110, 1);
      const innerMat = new THREE.MeshBasicMaterial({
        color: 0xfbbf24,
        wireframe: true,
        transparent: true,
        opacity: 0.04,
      });
      const innerMesh = new THREE.Mesh(innerGeo, innerMat);
      scene.add(innerMesh);

      // ── Particles scattered around the sphere surface ───────────────────────
      const NUM_PTS = 120;
      const ptPositions = new Float32Array(NUM_PTS * 3);
      const R = 175;
      for (let i = 0; i < NUM_PTS; i++) {
        // Uniform sphere surface distribution (Fibonacci sphere)
        const theta = Math.acos(1 - (2 * (i + 0.5)) / NUM_PTS);
        const phi   = Math.PI * (1 + Math.sqrt(5)) * i;
        ptPositions[i * 3]     = R * Math.sin(theta) * Math.cos(phi);
        ptPositions[i * 3 + 1] = R * Math.sin(theta) * Math.sin(phi);
        ptPositions[i * 3 + 2] = R * Math.cos(theta);
      }

      const ptGeo = new THREE.BufferGeometry();
      ptGeo.setAttribute('position', new THREE.BufferAttribute(ptPositions, 3));
      const ptMat = new THREE.PointsMaterial({
        color: 0xfbbf24, size: 3,
        sizeAttenuation: true,
        transparent: true, opacity: 0.45,
      });
      const pts = new THREE.Points(ptGeo, ptMat);
      scene.add(pts);

      // ── Resize ─────────────────────────────────────────────────────────────
      const onResize = () => {
        const w = el.clientWidth, h = el.clientHeight;
        renderer.setSize(w, h);
        camera.aspect = w / h;
        camera.updateProjectionMatrix();
      };
      window.addEventListener('resize', onResize);

      // ── Tick ───────────────────────────────────────────────────────────────
      let t = 0;

      const tick = () => {
        raf = requestAnimationFrame(tick);
        t += 0.004;

        // Counter-rotate the two shells for a gyroscope feel
        outerMesh.rotation.x = t * 0.3;
        outerMesh.rotation.y = t * 0.5;
        outerMesh.rotation.z = t * 0.15;

        innerMesh.rotation.x = -t * 0.4;
        innerMesh.rotation.y = t * 0.2;
        innerMesh.rotation.z = t * 0.35;

        // Particles ride with the outer shell
        pts.rotation.copy(outerMesh.rotation);

        renderer.render(scene, camera);
      };

      tick();

      doCleanup = () => {
        cancelAnimationFrame(raf);
        window.removeEventListener('resize', onResize);
        if (el.contains(renderer.domElement)) el.removeChild(renderer.domElement);
        renderer.dispose();
        outerGeo.dispose(); innerGeo.dispose(); ptGeo.dispose();
        outerMat.dispose(); innerMat.dispose(); ptMat.dispose();
      };
    });

    return () => {
      disposed = true;
      doCleanup?.();
    };
  }, []);

  return (
    <div
      ref={mountRef}
      className="absolute inset-0 pointer-events-none"
      aria-hidden="true"
    />
  );
}
