'use client';

import { useEffect, useRef } from 'react';

const COLS = 55;
const ROWS = 32;

export function WaveGrid() {
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
      renderer.setPixelRatio(Math.min(window.devicePixelRatio, 1.5));
      renderer.setSize(W, H);
      renderer.setClearColor(0x000000, 0);
      el.appendChild(renderer.domElement);

      // ── Camera slightly above, looking down at an angle ────────────────────
      const camera = new THREE.PerspectiveCamera(45, W / H, 1, 3000);
      camera.position.set(0, 320, 480);
      camera.lookAt(0, 0, 0);

      const scene = new THREE.Scene();

      // ── Wave geometry — PlaneGeometry laid flat, wireframe material ─────────
      const geo = new THREE.PlaneGeometry(1800, 900, COLS, ROWS);
      geo.rotateX(-Math.PI / 2);

      const posAttr = geo.getAttribute('position') as import('three').BufferAttribute;
      posAttr.setUsage(THREE.DynamicDrawUsage);

      const baseX = new Float32Array(posAttr.count);
      const baseZ = new Float32Array(posAttr.count);
      for (let i = 0; i < posAttr.count; i++) {
        baseX[i] = posAttr.getX(i);
        baseZ[i] = posAttr.getZ(i);
      }

      const dark = document.documentElement.classList.contains('dark');
      const mat = new THREE.MeshBasicMaterial({
        color: dark ? 0xffffff : 0x111111,
        wireframe: true,
        transparent: true,
        opacity: 0.042,
      });

      scene.add(new THREE.Mesh(geo, mat));

      // ── Theme observer — swap wireframe color when dark class changes ────────
      const themeObserver = new MutationObserver(() => {
        mat.color.setHex(
          document.documentElement.classList.contains('dark') ? 0xffffff : 0x111111
        );
      });
      themeObserver.observe(document.documentElement, {
        attributes: true, attributeFilter: ['class'],
      });

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
        t += 0.007;

        for (let i = 0; i < posAttr.count; i++) {
          const x = baseX[i];
          const z = baseZ[i];
          const y =
            Math.sin(x * 0.006 + t) * Math.cos(z * 0.009 + t * 0.65) * 38 +
            Math.sin(x * 0.012 + t * 1.3) * Math.sin(z * 0.006 + t * 0.5) * 18;
          posAttr.setY(i, y);
        }
        posAttr.needsUpdate = true;
        geo.computeVertexNormals();

        renderer.render(scene, camera);
      };

      tick();

      doCleanup = () => {
        cancelAnimationFrame(raf);
        themeObserver.disconnect();
        window.removeEventListener('resize', onResize);
        if (el.contains(renderer.domElement)) el.removeChild(renderer.domElement);
        renderer.dispose();
        geo.dispose();
        mat.dispose();
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
