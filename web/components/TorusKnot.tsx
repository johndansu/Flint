'use client';

import { useEffect, useRef } from 'react';

export function TorusKnot() {
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
      renderer.setClearColor(0x000000, 0);
      el.appendChild(renderer.domElement);

      const camera = new THREE.PerspectiveCamera(45, W / H, 1, 2000);
      camera.position.set(0, 0, 480);
      const scene = new THREE.Scene();

      const AMBER = 0xfbbf24;

      // ── (p=2, q=3) knot — the classic trefoil-like shape ─────────────────────
      // Outer: large, sparse, slowly rotating
      const outerGeo = new THREE.TorusKnotGeometry(130, 22, 256, 18, 2, 3);
      const outerMat = new THREE.MeshBasicMaterial({
        color: AMBER, wireframe: true, transparent: true, opacity: 0.045,
      });
      const outerMesh = new THREE.Mesh(outerGeo, outerMat);
      scene.add(outerMesh);

      // Inner: smaller (p=3,q=2) knot — different topology, counter-rotates
      const innerGeo = new THREE.TorusKnotGeometry(80, 14, 192, 14, 3, 2);
      const innerMat = new THREE.MeshBasicMaterial({
        color: AMBER, wireframe: true, transparent: true, opacity: 0.035,
      });
      const innerMesh = new THREE.Mesh(innerGeo, innerMat);
      scene.add(innerMesh);

      // ── Surface particles sampled from the outer knot's vertex positions ───────
      // We read the already-computed vertex positions instead of parameterising
      // the curve ourselves.
      const outerPos   = outerGeo.getAttribute('position');
      const STEP       = 8;                                    // sample every N-th vertex
      const ptCount    = Math.floor(outerPos.count / STEP);
      const ptBuf      = new Float32Array(ptCount * 3);
      for (let i = 0; i < ptCount; i++) {
        ptBuf[i*3]   = outerPos.getX(i * STEP);
        ptBuf[i*3+1] = outerPos.getY(i * STEP);
        ptBuf[i*3+2] = outerPos.getZ(i * STEP);
      }
      const ptGeo = new THREE.BufferGeometry();
      ptGeo.setAttribute('position', new THREE.BufferAttribute(ptBuf, 3));
      const ptMat = new THREE.PointsMaterial({
        color: AMBER, size: 3, sizeAttenuation: true,
        transparent: true, opacity: 0.60,
      });
      const pts = new THREE.Points(ptGeo, ptMat);
      scene.add(pts);

      // ── Thin bright spine — a single tube outline for the outer knot ──────────
      // Sample tube-centre positions (every vertex of the path curve, not the tube)
      const spineGeo = new THREE.TorusKnotGeometry(130, 0.8, 512, 4, 2, 3);
      const spineMat = new THREE.MeshBasicMaterial({
        color: AMBER, wireframe: true, transparent: true, opacity: 0.20,
      });
      scene.add(new THREE.Mesh(spineGeo, spineMat));

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

      // ── Tick ──────────────────────────────────────────────────────────────────
      let t = 0;
      const tick = () => {
        raf = requestAnimationFrame(tick);
        t += 0.0035;

        // Outer knot: slow rotation in all axes for depth
        outerMesh.rotation.x =  t * 0.30;
        outerMesh.rotation.y =  t * 0.45;
        outerMesh.rotation.z =  t * 0.18;

        // Inner knot: counter-rotates for a gyroscope feel
        innerMesh.rotation.x = -t * 0.38;
        innerMesh.rotation.y =  t * 0.22;
        innerMesh.rotation.z = -t * 0.50;

        // Surface particles ride with the outer mesh
        pts.rotation.copy(outerMesh.rotation);

        // Spine follows outer rotation too
        scene.children[3].rotation.copy(outerMesh.rotation);

        // Smooth camera parallax toward mouse
        camera.position.x += (mx * 60 - camera.position.x) * 0.03;
        camera.position.y += (-my * 35 - camera.position.y) * 0.03;
        camera.lookAt(0, 0, 0);

        renderer.render(scene, camera);
      };
      tick();

      doCleanup = () => {
        cancelAnimationFrame(raf);
        window.removeEventListener('mousemove', onMouse);
        window.removeEventListener('resize', onResize);
        if (el.contains(renderer.domElement)) el.removeChild(renderer.domElement);
        renderer.dispose();
        outerGeo.dispose(); outerMat.dispose();
        innerGeo.dispose(); innerMat.dispose();
        ptGeo.dispose();    ptMat.dispose();
        spineGeo.dispose(); spineMat.dispose();
      };
    });

    return () => { disposed = true; doCleanup?.(); };
  }, []);

  return (
    <div ref={mountRef} className="absolute inset-0 pointer-events-none" aria-hidden="true" />
  );
}
