'use client';

import { useEffect, useRef } from 'react';

const N     = 90;
const DIST  = 170;
const SPEED = 0.20;

export function HeroCanvas() {
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
      const renderer = new THREE.WebGLRenderer({ alpha: true, antialias: false });
      renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));
      renderer.setSize(W, H);
      renderer.setClearColor(0x000000, 0);
      el.appendChild(renderer.domElement);

      // ── PerspectiveCamera — depth makes 2D particle nets feel 3D ───────────
      const camera = new THREE.PerspectiveCamera(55, W / H, 1, 3000);
      camera.position.set(0, 0, 600);

      const scene = new THREE.Scene();

      // ── Particles ──────────────────────────────────────────────────────────
      const spread = Math.min(W * 0.9, 700);
      const pos    = new Float32Array(N * 3);
      const vx     = new Float32Array(N);
      const vy     = new Float32Array(N);
      const vz     = new Float32Array(N);

      for (let i = 0; i < N; i++) {
        pos[i * 3]     = (Math.random() - 0.5) * spread;
        pos[i * 3 + 1] = (Math.random() - 0.5) * spread * (H / W);
        pos[i * 3 + 2] = (Math.random() - 0.5) * 220;
        vx[i] = (Math.random() - 0.5) * SPEED * 2;
        vy[i] = (Math.random() - 0.5) * SPEED * 2;
        vz[i] = (Math.random() - 0.5) * SPEED * 0.5;
      }

      const ptGeo  = new THREE.BufferGeometry();
      const ptAttr = new THREE.BufferAttribute(pos, 3);
      ptAttr.setUsage(THREE.DynamicDrawUsage);
      ptGeo.setAttribute('position', ptAttr);

      const dark = document.documentElement.classList.contains('dark');
      const themeColor = dark ? 0xffffff : 0x111111;

      const ptMat = new THREE.PointsMaterial({
        color: themeColor, size: 3,
        sizeAttenuation: true,
        transparent: true, opacity: 0.28,
      });
      scene.add(new THREE.Points(ptGeo, ptMat));

      // ── Lines ──────────────────────────────────────────────────────────────
      const maxSeg = (N * (N - 1)) / 2;
      const lnPos  = new Float32Array(maxSeg * 6);
      const lnGeo  = new THREE.BufferGeometry();
      const lnAttr = new THREE.BufferAttribute(lnPos, 3);
      lnAttr.setUsage(THREE.DynamicDrawUsage);
      lnGeo.setAttribute('position', lnAttr);

      const lnMat = new THREE.LineBasicMaterial({
        color: themeColor, transparent: true, opacity: 0.055,
      });
      scene.add(new THREE.LineSegments(lnGeo, lnMat));

      // ── Theme observer — swap colors when dark class is toggled ─────────────
      const themeObserver = new MutationObserver(() => {
        const c = document.documentElement.classList.contains('dark') ? 0xffffff : 0x111111;
        ptMat.color.setHex(c);
        lnMat.color.setHex(c);
      });
      themeObserver.observe(document.documentElement, {
        attributes: true, attributeFilter: ['class'],
      });

      // ── Mouse parallax ─────────────────────────────────────────────────────
      let mx = 0, my = 0;
      const onMouse = (e: MouseEvent) => {
        mx = (e.clientX / window.innerWidth  - 0.5) * 2;
        my = (e.clientY / window.innerHeight - 0.5) * 2;
      };
      window.addEventListener('mousemove', onMouse);

      // ── Scroll-driven depth — camera recedes as user scrolls away ───────────
      let targetZ = 600;
      const onScroll = () => { targetZ = 600 + window.scrollY * 0.38; };
      window.addEventListener('scroll', onScroll, { passive: true });

      // ── Resize ─────────────────────────────────────────────────────────────
      const onResize = () => {
        const w = el.clientWidth, h = el.clientHeight;
        renderer.setSize(w, h);
        camera.aspect = w / h;
        camera.updateProjectionMatrix();
      };
      window.addEventListener('resize', onResize);

      // ── Tick ───────────────────────────────────────────────────────────────
      const DIST2   = DIST * DIST;
      const halfX   = spread / 2;
      const halfY   = (spread * (H / W)) / 2;
      const halfZ   = 120;

      const tick = () => {
        raf = requestAnimationFrame(tick);

        camera.position.x += (mx * 50 - camera.position.x) * 0.025;
        camera.position.y += (-my * 25 - camera.position.y) * 0.025;
        camera.position.z += (targetZ - camera.position.z) * 0.06;
        camera.lookAt(0, 0, 0);

        for (let i = 0; i < N; i++) {
          pos[i * 3]     += vx[i];
          pos[i * 3 + 1] += vy[i];
          pos[i * 3 + 2] += vz[i];
          if (Math.abs(pos[i * 3])     > halfX) vx[i] *= -1;
          if (Math.abs(pos[i * 3 + 1]) > halfY) vy[i] *= -1;
          if (Math.abs(pos[i * 3 + 2]) > halfZ) vz[i] *= -1;
        }
        ptAttr.needsUpdate = true;

        let li = 0;
        for (let i = 0; i < N; i++) {
          const ix = pos[i * 3], iy = pos[i * 3 + 1], iz = pos[i * 3 + 2];
          for (let j = i + 1; j < N; j++) {
            const dx = ix - pos[j * 3];
            const dy = iy - pos[j * 3 + 1];
            const dz = iz - pos[j * 3 + 2];
            if (dx * dx + dy * dy + dz * dz < DIST2) {
              lnPos[li++] = ix;          lnPos[li++] = iy;             lnPos[li++] = iz;
              lnPos[li++] = pos[j * 3]; lnPos[li++] = pos[j * 3 + 1]; lnPos[li++] = pos[j * 3 + 2];
            }
          }
        }
        lnAttr.needsUpdate = true;
        lnGeo.setDrawRange(0, li / 3);

        renderer.render(scene, camera);
      };

      tick();

      doCleanup = () => {
        cancelAnimationFrame(raf);
        themeObserver.disconnect();
        window.removeEventListener('mousemove', onMouse);
        window.removeEventListener('scroll', onScroll);
        window.removeEventListener('resize', onResize);
        if (el.contains(renderer.domElement)) el.removeChild(renderer.domElement);
        renderer.dispose();
        ptGeo.dispose(); lnGeo.dispose();
        ptMat.dispose(); lnMat.dispose();
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
