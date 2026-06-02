'use client';

import { useEffect, useRef } from 'react';

// Magnetic dipole field lines: r(θ) = r_max · sin²θ
// Creates the aurora/compass shape — nested luminous loops pole-to-pole.

const R_MAX = [52, 85, 122, 164, 210, 260];   // equatorial crossing radii per family
const N_PHI = 10;                               // azimuth copies per family
const PTS   = 92;                               // samples per field line
const EPS   = 0.072;                            // avoid pole singularity

export function MagneticField() {
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

      const renderer = new THREE.WebGLRenderer({ alpha: true, antialias: false });
      renderer.setPixelRatio(Math.min(window.devicePixelRatio, 1.5));
      renderer.setSize(W, H);
      renderer.setClearColor(0x000000, 0);
      el.appendChild(renderer.domElement);

      const camera = new THREE.PerspectiveCamera(50, W / H, 1, 3000);
      camera.position.set(0, 60, 450);
      camera.lookAt(0, 0, 0);

      const scene  = new THREE.Scene();
      const dark   = document.documentElement.classList.contains('dark');
      const N_FAMS = R_MAX.length;
      const N_LINES = N_FAMS * N_PHI;
      const N_PTS   = N_LINES * PTS;
      const N_SEGS  = N_LINES * (PTS - 1);

      const ptPos   = new Float32Array(N_PTS  * 3);
      const ptColor = new Float32Array(N_PTS  * 3);
      const lnPos   = new Float32Array(N_SEGS * 6);   // 2 verts × 3 floats per segment

      // ── Build geometry ──────────────────────────────────────────────────────
      const fillColor = (ptIdx: number, theta: number, ri: number, d: boolean) => {
        const poleProx  = Math.abs(Math.cos(theta));          // 1 at poles, 0 at equator
        const innerBias = 1 - (ri / N_FAMS) * 0.42;          // inner families brighter
        const br = d
          ? (0.10 + poleProx * 0.40) * innerBias
          : (0.06 + poleProx * 0.30) * innerBias;
        ptColor[ptIdx * 3]     = d ? br         : br * 0.90;
        ptColor[ptIdx * 3 + 1] = d ? br * 0.60  : br * 0.44;
        ptColor[ptIdx * 3 + 2] = d ? br * 0.05  : 0;
      };

      let ptIdx = 0;
      let segIdx = 0;

      for (let ri = 0; ri < N_FAMS; ri++) {
        const rmax = R_MAX[ri];

        for (let fi = 0; fi < N_PHI; fi++) {
          const phi = (fi / N_PHI) * Math.PI * 2;
          let prevX = 0, prevY = 0, prevZ = 0;

          for (let k = 0; k < PTS; k++) {
            const theta = EPS + (k / (PTS - 1)) * (Math.PI - 2 * EPS);
            const r     = rmax * Math.sin(theta) ** 2;
            const rho   = r * Math.sin(theta);
            const x     = rho * Math.cos(phi);
            const y     = r   * Math.cos(theta);
            const z     = rho * Math.sin(phi);

            ptPos[ptIdx * 3]     = x;
            ptPos[ptIdx * 3 + 1] = y;
            ptPos[ptIdx * 3 + 2] = z;
            fillColor(ptIdx, theta, ri, dark);

            if (k > 0) {
              lnPos[segIdx * 6]     = prevX;
              lnPos[segIdx * 6 + 1] = prevY;
              lnPos[segIdx * 6 + 2] = prevZ;
              lnPos[segIdx * 6 + 3] = x;
              lnPos[segIdx * 6 + 4] = y;
              lnPos[segIdx * 6 + 5] = z;
              segIdx++;
            }

            prevX = x; prevY = y; prevZ = z;
            ptIdx++;
          }
        }
      }

      // ── Points ──────────────────────────────────────────────────────────────
      const ptGeo   = new THREE.BufferGeometry();
      const colAttr = new THREE.BufferAttribute(ptColor, 3);
      ptGeo.setAttribute('position', new THREE.BufferAttribute(ptPos, 3));
      ptGeo.setAttribute('color',    colAttr);

      const ptMat = new THREE.PointsMaterial({
        size: 2.2, sizeAttenuation: true,
        vertexColors: true, transparent: true, opacity: 0.55,
      });
      const points = new THREE.Points(ptGeo, ptMat);

      // ── Lines (uniform amber, low opacity — the "glow skeleton") ─────────────
      const lnGeo = new THREE.BufferGeometry();
      lnGeo.setAttribute('position', new THREE.BufferAttribute(lnPos, 3));
      const lnMat = new THREE.LineBasicMaterial({
        color: dark ? 0xfbbf24 : 0xd97706,
        transparent: true, opacity: 0.08,
      });
      const segs = new THREE.LineSegments(lnGeo, lnMat);

      const group = new THREE.Group();
      group.add(segs);
      group.add(points);
      scene.add(group);

      // ── Theme ────────────────────────────────────────────────────────────────
      const themeObs = new MutationObserver(() => {
        const d = document.documentElement.classList.contains('dark');
        lnMat.color.setHex(d ? 0xfbbf24 : 0xd97706);
        let i = 0;
        for (let ri = 0; ri < N_FAMS; ri++) {
          for (let fi = 0; fi < N_PHI; fi++) {
            for (let k = 0; k < PTS; k++) {
              fillColor(i, EPS + (k / (PTS - 1)) * (Math.PI - 2 * EPS), ri, d);
              i++;
            }
          }
        }
        colAttr.needsUpdate = true;
      });
      themeObs.observe(document.documentElement, { attributes: true, attributeFilter: ['class'] });

      // ── Scroll boost ─────────────────────────────────────────────────────────
      let scrollBoost = 0;
      let lastY = window.scrollY;
      const onScroll = () => {
        scrollBoost += Math.abs(window.scrollY - lastY) * 0.7;
        lastY = window.scrollY;
      };
      window.addEventListener('scroll', onScroll, { passive: true });

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

      // ── Tick ─────────────────────────────────────────────────────────────────
      let rotY = 0;
      let camT = 0;
      const tick = () => {
        raf = requestAnimationFrame(tick);
        camT += 0.001;

        const angVel = 0.0005 + scrollBoost * 0.003;
        scrollBoost *= 0.88;
        rotY += angVel;

        group.rotation.y = rotY;
        group.rotation.z = Math.sin(rotY * 0.7) * 0.13;   // gentle precession

        // Slow camera orbit + mouse lean
        const camTX = Math.cos(camT * 0.35) * 38 + mx * 52;
        const camTY = Math.sin(camT * 0.28) * 22 - my * 36 + 60;
        camera.position.x += (camTX - camera.position.x) * 0.022;
        camera.position.y += (camTY - camera.position.y) * 0.022;
        camera.lookAt(0, 0, 0);

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
        ptGeo.dispose(); ptMat.dispose();
        lnGeo.dispose(); lnMat.dispose();
      };
    });

    return () => { disposed = true; doCleanup?.(); };
  }, []);

  return (
    <div ref={mountRef} className="absolute inset-0 pointer-events-none" aria-hidden="true" />
  );
}
