'use client';

import { useEffect, useRef } from 'react';

export function OrbitalRings() {
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

      // ── Renderer ─────────────────────────────────────────────────────────────
      const renderer = new THREE.WebGLRenderer({ alpha: true, antialias: true });
      renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));
      renderer.setSize(W, H);
      renderer.setClearColor(0x000000, 0);
      el.appendChild(renderer.domElement);

      const camera = new THREE.PerspectiveCamera(50, W / H, 1, 2000);
      camera.position.set(0, 0, 600);
      const scene = new THREE.Scene();

      const AMBER = 0xfbbf24;

      // ── Ring factory ──────────────────────────────────────────────────────────
      // Each ring: outer spinGroup (animated Y) → inner tiltGroup (fixed X/Z) → geometry.
      // This avoids gimbal lock — tilt baked in child, orbit driven in parent.
      const makeRing = (
        R: number, seg: number,
        tiltX: number, tiltZ: number,
        lineOp: number, dotOp: number, dotCount: number
      ) => {
        const spinGroup = new THREE.Group();
        const tiltGroup = new THREE.Group();
        tiltGroup.rotation.x = tiltX;
        tiltGroup.rotation.z = tiltZ;
        spinGroup.add(tiltGroup);

        // Closed ring guide line
        const pts: import('three').Vector3[] = [];
        for (let i = 0; i <= seg; i++) {
          const a = (i / seg) * Math.PI * 2;
          pts.push(new THREE.Vector3(Math.cos(a) * R, Math.sin(a) * R, 0));
        }
        const lineGeo = new THREE.BufferGeometry().setFromPoints(pts);
        const lineMat = new THREE.LineBasicMaterial({
          color: AMBER, transparent: true, opacity: lineOp,
        });
        tiltGroup.add(new THREE.Line(lineGeo, lineMat));

        // Dots orbiting along the ring path
        const dotPos = new Float32Array(dotCount * 3);
        for (let i = 0; i < dotCount; i++) {
          const a = (i / dotCount) * Math.PI * 2;
          dotPos[i*3]   = Math.cos(a) * R;
          dotPos[i*3+1] = Math.sin(a) * R;
          dotPos[i*3+2] = 0;
        }
        const dotGeo = new THREE.BufferGeometry();
        dotGeo.setAttribute('position', new THREE.BufferAttribute(dotPos, 3));
        const dotMat = new THREE.PointsMaterial({
          color: AMBER, size: 3.5, sizeAttenuation: true,
          transparent: true, opacity: dotOp,
        });
        tiltGroup.add(new THREE.Points(dotGeo, dotMat));

        return { spinGroup, lineGeo, lineMat, dotGeo, dotMat };
      };

      // Three rings with different radii, tilts, and spin speeds
      const ring1 = makeRing(230, 128, Math.PI * 0.14, 0,              0.13, 0.55, 28);
      const ring2 = makeRing(170, 100, Math.PI * 0.55, 0,              0.09, 0.45, 20);
      const ring3 = makeRing(295, 150, Math.PI * 0.80, Math.PI * 0.35, 0.06, 0.28, 36);

      // World group — receives mouse parallax, contains all spin groups
      const world = new THREE.Group();
      world.add(ring1.spinGroup, ring2.spinGroup, ring3.spinGroup);
      scene.add(world);

      // ── Central core ─────────────────────────────────────────────────────────
      const coreGeo  = new THREE.IcosahedronGeometry(26, 2);
      const coreMat  = new THREE.MeshBasicMaterial({
        color: AMBER, wireframe: true, transparent: true, opacity: 0.22,
      });
      const coreMesh = new THREE.Mesh(coreGeo, coreMat);
      scene.add(coreMesh);

      // ── Mouse parallax ────────────────────────────────────────────────────────
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
        t += 0.0042;

        // Each ring orbits independently (spinGroup.rotation.y = orbital angle)
        ring1.spinGroup.rotation.y =  t * 0.42;
        ring2.spinGroup.rotation.y = -t * 0.58;
        ring3.spinGroup.rotation.y =  t * 0.24;

        // World drifts gently with mouse (smooth follow)
        world.rotation.x += (-my * 0.22 - world.rotation.x) * 0.04;
        world.rotation.y += ( mx * 0.28 - world.rotation.y) * 0.04;

        // Core spins independently
        coreMesh.rotation.x = t * 0.55;
        coreMesh.rotation.y = t * 0.85;

        renderer.render(scene, camera);
      };
      tick();

      doCleanup = () => {
        cancelAnimationFrame(raf);
        window.removeEventListener('mousemove', onMouse);
        window.removeEventListener('resize', onResize);
        if (el.contains(renderer.domElement)) el.removeChild(renderer.domElement);
        renderer.dispose();
        [ring1, ring2, ring3].forEach(r => {
          r.lineGeo.dispose(); r.lineMat.dispose();
          r.dotGeo.dispose();  r.dotMat.dispose();
        });
        coreGeo.dispose(); coreMat.dispose();
      };
    });

    return () => { disposed = true; doCleanup?.(); };
  }, []);

  return (
    <div ref={mountRef} className="absolute inset-0 pointer-events-none" aria-hidden="true" />
  );
}
