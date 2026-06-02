'use client';

import { useEffect, useRef } from 'react';

export function MorphSphere() {
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
      camera.position.set(0, 0, 440);
      const scene = new THREE.Scene();

      const AMBER = 0xfbbf24;

      // ── Outer morphing shell (detail=3 = 1280 faces / 3840 buffer verts) ────
      const geo     = new THREE.IcosahedronGeometry(155, 3);
      const posAttr = geo.getAttribute('position') as import('three').BufferAttribute;
      posAttr.setUsage(THREE.DynamicDrawUsage);

      // Snapshot base positions before any morphing
      const base = new Float32Array(posAttr.array.length);
      for (let i = 0; i < posAttr.array.length; i++) {
        base[i] = (posAttr.array as Float32Array)[i];
      }

      const mat = new THREE.MeshBasicMaterial({
        color: AMBER, wireframe: true, transparent: true, opacity: 0.055,
      });
      const mesh = new THREE.Mesh(geo, mat);
      scene.add(mesh);

      // ── Inner static shell (gives sense of depth inside the orb) ─────────────
      const innerGeo = new THREE.IcosahedronGeometry(95, 2);
      const innerMat = new THREE.MeshBasicMaterial({
        color: AMBER, wireframe: true, transparent: true, opacity: 0.022,
      });
      const innerMesh = new THREE.Mesh(innerGeo, innerMat);
      scene.add(innerMesh);

      // ── Surface dot particles ────────────────────────────────────────────────
      const dotGeo = new THREE.IcosahedronGeometry(157, 2);
      const dotMat = new THREE.PointsMaterial({
        color: AMBER, size: 2.5, sizeAttenuation: true,
        transparent: true, opacity: 0.30,
      });
      scene.add(new THREE.Points(dotGeo, dotMat));

      // ── Resize ───────────────────────────────────────────────────────────────
      const onResize = () => {
        const w = el.clientWidth, h = el.clientHeight;
        renderer.setSize(w, h);
        camera.aspect = w / h;
        camera.updateProjectionMatrix();
      };
      window.addEventListener('resize', onResize);

      // ── Tick — organic vertex displacement along surface normals ─────────────
      let t = 0;
      const tick = () => {
        raf = requestAnimationFrame(tick);
        t += 0.007;

        const count = posAttr.count;
        for (let i = 0; i < count; i++) {
          const bx = base[i*3], by = base[i*3+1], bz = base[i*3+2];
          const len = Math.sqrt(bx*bx + by*by + bz*bz) || 1;
          const nx  = bx/len, ny = by/len, nz = bz/len;

          // Two overlapping sine waves create organic pulsing
          const d =
            Math.sin(bx * 0.022 + t * 1.1) * Math.cos(by * 0.022 + t * 0.8) * 20 +
            Math.cos(bz * 0.028 + t * 0.6) * Math.sin(bx * 0.015 + t * 0.9) * 12;

          const r = 155 + d;
          posAttr.setXYZ(i, nx*r, ny*r, nz*r);
        }
        posAttr.needsUpdate = true;

        mesh.rotation.y      =  t * 0.18;
        mesh.rotation.x      =  t * 0.10;
        innerMesh.rotation.y = -t * 0.24;
        innerMesh.rotation.z =  t * 0.13;

        renderer.render(scene, camera);
      };
      tick();

      doCleanup = () => {
        cancelAnimationFrame(raf);
        window.removeEventListener('resize', onResize);
        if (el.contains(renderer.domElement)) el.removeChild(renderer.domElement);
        renderer.dispose();
        geo.dispose(); innerGeo.dispose(); dotGeo.dispose();
        mat.dispose(); innerMat.dispose(); dotMat.dispose();
      };
    });

    return () => { disposed = true; doCleanup?.(); };
  }, []);

  return (
    <div ref={mountRef} className="absolute inset-0 pointer-events-none" aria-hidden="true" />
  );
}
