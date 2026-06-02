'use client';

import { useEffect, useRef } from 'react';

// 3D neural network — nodes connected by edges, bright pulses
// travel along edges like action potentials through synapses.

const N_NODES  = 52;
const SPHERE_R = 148;
const CONN_R   = 88;
const MAX_DEG  = 5;

export function NeuralWeb() {
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

      const camera = new THREE.PerspectiveCamera(50, W / H, 1, 2000);
      camera.position.set(0, 0, 420);

      const scene = new THREE.Scene();
      const dark  = document.documentElement.classList.contains('dark');

      // ── Node positions — uniform in sphere volume ────────────────────────────
      const nodePos = new Float32Array(N_NODES * 3);
      for (let i = 0; i < N_NODES; i++) {
        const r     = SPHERE_R * Math.cbrt(Math.random());
        const theta = Math.acos(1 - 2 * Math.random());
        const phi   = Math.PI * 2 * Math.random();
        nodePos[i * 3]     = r * Math.sin(theta) * Math.cos(phi);
        nodePos[i * 3 + 1] = r * Math.cos(theta);
        nodePos[i * 3 + 2] = r * Math.sin(theta) * Math.sin(phi);
      }

      // ── Build edges ──────────────────────────────────────────────────────────
      const edges: [number, number][] = [];
      const deg = new Array<number>(N_NODES).fill(0);
      for (let i = 0; i < N_NODES; i++) {
        for (let j = i + 1; j < N_NODES; j++) {
          if (deg[i] >= MAX_DEG || deg[j] >= MAX_DEG) continue;
          const dx = nodePos[i*3]   - nodePos[j*3];
          const dy = nodePos[i*3+1] - nodePos[j*3+1];
          const dz = nodePos[i*3+2] - nodePos[j*3+2];
          if (Math.sqrt(dx*dx + dy*dy + dz*dz) < CONN_R) {
            edges.push([i, j]);
            deg[i]++; deg[j]++;
          }
        }
      }
      const NE = edges.length;

      // ── Static edge lines ────────────────────────────────────────────────────
      const edgePos = new Float32Array(NE * 6);
      for (let i = 0; i < NE; i++) {
        const [a, b] = edges[i];
        edgePos[i*6]   = nodePos[a*3];   edgePos[i*6+1] = nodePos[a*3+1]; edgePos[i*6+2] = nodePos[a*3+2];
        edgePos[i*6+3] = nodePos[b*3];   edgePos[i*6+4] = nodePos[b*3+1]; edgePos[i*6+5] = nodePos[b*3+2];
      }
      const edgeGeo = new THREE.BufferGeometry();
      edgeGeo.setAttribute('position', new THREE.BufferAttribute(edgePos, 3));
      const edgeMat = new THREE.LineBasicMaterial({
        color: dark ? 0xfbbf24 : 0xd97706, transparent: true, opacity: 0.12,
      });
      scene.add(new THREE.LineSegments(edgeGeo, edgeMat));

      // ── Node dots ────────────────────────────────────────────────────────────
      const nodeGeo = new THREE.BufferGeometry();
      nodeGeo.setAttribute('position', new THREE.BufferAttribute(nodePos.slice(), 3));
      const nodeMat = new THREE.PointsMaterial({
        color: dark ? 0xfbbf24 : 0xd97706, size: 4.5, sizeAttenuation: true,
        transparent: true, opacity: 0.65,
      });
      scene.add(new THREE.Points(nodeGeo, nodeMat));

      // ── Pulse particles (one per edge, travels nodeA → nodeB) ────────────────
      const pulsePosArr  = new Float32Array(NE * 3);
      const pulseGeo     = new THREE.BufferGeometry();
      const pulsePosAttr = new THREE.BufferAttribute(pulsePosArr, 3);
      pulsePosAttr.setUsage(THREE.DynamicDrawUsage);
      pulseGeo.setAttribute('position', pulsePosAttr);
      const pulseMat = new THREE.PointsMaterial({
        color: 0xfbbf24, size: 5, sizeAttenuation: true, transparent: true, opacity: 0.90,
      });
      scene.add(new THREE.Points(pulseGeo, pulseMat));

      // Each pulse: a phase 0→1, speed, and stagger offset
      const phases  = new Float32Array(NE);
      const speeds  = new Float32Array(NE);
      const offsets = new Float32Array(NE);
      for (let i = 0; i < NE; i++) {
        phases[i]  = Math.random();
        speeds[i]  = 0.005 + Math.random() * 0.008;
        offsets[i] = Math.random();  // pulse is "visible" only when active window passes
      }

      // ── Group for rotation ────────────────────────────────────────────────────
      // Move all objects into a group so they rotate together
      const group = new THREE.Group();
      scene.children.slice().forEach(c => { scene.remove(c); group.add(c); });
      scene.add(group);

      // ── Theme ────────────────────────────────────────────────────────────────
      const themeObs = new MutationObserver(() => {
        const d = document.documentElement.classList.contains('dark');
        const c = d ? 0xfbbf24 : 0xd97706;
        edgeMat.color.setHex(c);
        nodeMat.color.setHex(c);
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

      // ── Tick ─────────────────────────────────────────────────────────────────
      let t = 0;
      const tick = () => {
        raf = requestAnimationFrame(tick);
        t += 0.003;

        // Advance pulses
        for (let i = 0; i < NE; i++) {
          phases[i] += speeds[i];
          if (phases[i] >= 1) phases[i] -= 1;

          // Pulse is visible only in a 20% window of the cycle
          const visible = phases[i] < 0.20;
          if (visible) {
            const frac = phases[i] / 0.20;   // 0→1 within the visible window
            const [a, b] = edges[i];
            pulsePosArr[i*3]   = nodePos[a*3]   + (nodePos[b*3]   - nodePos[a*3])   * frac;
            pulsePosArr[i*3+1] = nodePos[a*3+1] + (nodePos[b*3+1] - nodePos[a*3+1]) * frac;
            pulsePosArr[i*3+2] = nodePos[a*3+2] + (nodePos[b*3+2] - nodePos[a*3+2]) * frac;
          } else {
            // Park far away — invisible
            pulsePosArr[i*3] = 1e6; pulsePosArr[i*3+1] = 1e6; pulsePosArr[i*3+2] = 1e6;
          }
        }
        pulsePosAttr.needsUpdate = true;

        // Slow group rotation
        group.rotation.y =  t * 0.18;
        group.rotation.x =  Math.sin(t * 0.12) * 0.14;

        // Camera mouse parallax
        camera.position.x += (mx * 52 - camera.position.x) * 0.025;
        camera.position.y += (-my * 35 - camera.position.y) * 0.025;
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
        nodeGeo.dispose(); nodeMat.dispose();
        pulseGeo.dispose(); pulseMat.dispose();
      };
    });

    return () => { disposed = true; doCleanup?.(); };
  }, []);

  return (
    <div ref={mountRef} className="absolute inset-0 pointer-events-none" aria-hidden="true" />
  );
}
