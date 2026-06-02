'use client';

import { useEffect, useRef } from 'react';

const RINGS        = 42;
const PTS_PER_RING = 80;
const RING_R       = 195;
const RING_SPACING = 68;
const TUNNEL_LEN   = RINGS * RING_SPACING;
const BASE_SPEED   = 1.5;

export function ParticleTunnel() {
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

      // Camera sits at origin looking down -Z — every ring ahead creates infinite
      // depth. Perspective makes far rings small and near rings huge (parallax).
      const camera = new THREE.PerspectiveCamera(82, W / H, 5, 5000);
      camera.position.set(0, 0, 0);

      const scene = new THREE.Scene();
      const dark  = document.documentElement.classList.contains('dark');

      // ── Per-ring geometry parameters (fixed for life of component) ─────────────
      const ringR  = new Float32Array(RINGS);
      const ringOX = new Float32Array(RINGS);
      const ringOY = new Float32Array(RINGS);
      const ringTw = new Float32Array(RINGS);
      const ringZ  = new Float32Array(RINGS);

      for (let r = 0; r < RINGS; r++) {
        ringR[r]  = RING_R * (0.80 + Math.random() * 0.40);
        ringOX[r] = (Math.random() - 0.5) * 45;
        ringOY[r] = (Math.random() - 0.5) * 45;
        ringTw[r] = r * 0.42;
        ringZ[r]  = -(r * RING_SPACING) - RING_SPACING;  // spread from -SPACING to -TUNNEL_LEN
      }

      // ── Particle buffer ────────────────────────────────────────────────────────
      const N         = RINGS * PTS_PER_RING;
      const positions = new Float32Array(N * 3);

      const fillRing = (r: number) => {
        for (let p = 0; p < PTS_PER_RING; p++) {
          const i     = r * PTS_PER_RING + p;
          const angle = (p / PTS_PER_RING) * Math.PI * 2 + ringTw[r];
          positions[i*3]   = Math.cos(angle) * ringR[r] + ringOX[r];
          positions[i*3+1] = Math.sin(angle) * ringR[r] + ringOY[r];
          positions[i*3+2] = ringZ[r];
        }
      };

      for (let r = 0; r < RINGS; r++) fillRing(r);

      const geo     = new THREE.BufferGeometry();
      const posAttr = new THREE.BufferAttribute(positions, 3);
      posAttr.setUsage(THREE.DynamicDrawUsage);
      geo.setAttribute('position', posAttr);

      const mat = new THREE.PointsMaterial({
        color:           dark ? 0xfbbf24 : 0xd97706,
        size:            3.5,
        sizeAttenuation: true,   // near rings huge, far rings tiny — pure 3-D depth cue
        transparent:     true,
        opacity:         0.62,
      });
      scene.add(new THREE.Points(geo, mat));

      // ── Theme ─────────────────────────────────────────────────────────────────
      const themeObs = new MutationObserver(() => {
        mat.color.setHex(
          document.documentElement.classList.contains('dark') ? 0xfbbf24 : 0xd97706
        );
      });
      themeObs.observe(document.documentElement, { attributes: true, attributeFilter: ['class'] });

      // ── Scroll — speed spikes when the user scrolls through this section ───────
      let scrollBoost = 0;
      let lastY       = window.scrollY;
      const onScroll  = () => {
        scrollBoost += Math.abs(window.scrollY - lastY) * 0.55;
        lastY = window.scrollY;
      };
      window.addEventListener('scroll', onScroll, { passive: true });

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
      const tick = () => {
        raf = requestAnimationFrame(tick);
        const speed = BASE_SPEED + scrollBoost;
        scrollBoost *= 0.88;    // smooth decay

        // Advance every ring toward the camera (Z → 0)
        for (let r = 0; r < RINGS; r++) {
          ringZ[r] += speed;
          if (ringZ[r] > 0) {
            // Ring passed behind camera — teleport to far end (invisible transition)
            ringZ[r] -= TUNNEL_LEN;
          }
          // Update only the Z component — XY shape stays constant per ring
          const base = r * PTS_PER_RING;
          for (let p = 0; p < PTS_PER_RING; p++) {
            positions[(base + p) * 3 + 2] = ringZ[r];
          }
        }
        posAttr.needsUpdate = true;

        // Camera sway follows mouse so the tunnel "steers"
        camera.position.x += (mx * 28 - camera.position.x) * 0.04;
        camera.position.y += (-my * 18 - camera.position.y) * 0.04;
        camera.lookAt(
          camera.position.x * 0.06,
          camera.position.y * 0.06,
          -600
        );

        renderer.render(scene, camera);
      };
      tick();

      doCleanup = () => {
        cancelAnimationFrame(raf);
        themeObs.disconnect();
        window.removeEventListener('scroll', onScroll);
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
