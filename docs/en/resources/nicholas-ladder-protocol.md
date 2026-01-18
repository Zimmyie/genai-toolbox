---
title: "Nicholas Ladder Protocol"
type: docs
weight: 50
description: >
  A prereg-ready protocol for testing the Nicholas Ladder acoustic perception hypothesis.
---

# Nicholas Ladder Protocol

This document ships three plug-and-play pieces: a one-page prereg summary,
stimulus-generation pseudocode, and a free-draw feature-extractor spec. It is
prereg-ready, falsifiable, and reviewer-proof.

## One-page prereg summary (drop-in)

**Title**

Human–Room Coupling via Modal Symmetry Classes (“Nicholas Ladder”): Ordered vs
Shuffled Mapping Across Rooms

**Core claim (Law)**

Low-order acoustic mode families in a cavity (radial, 2-lobe, 3-lobe,
orthogonal) elicit robust, naïve perceptual Gestalts (○, vesica, △, □). Higher
families (5/6-ray) act as trainable attractors; a broadband, mixed-mode regime
yields a spiral/scale-free percept.

**Formalization (minimal math)**

Solve \(\nabla^2\psi_n + k_n^2\psi_n = 0\) on \(\Omega\); define nondimensional
\(\mathrm{He}_n = k_n\cdot L\). Map eigenfunction \(\psi_n\) to symmetry class
\(\sigma(\psi_n)\in\{\text{radial}, 2\text{-lobe}, 3\text{-lobe},\text{orthogonal}, 5\text{-ray}, 6\text{-ray},\text{scale-free}\}\). Let report
\(Y\in\{\bigcirc, \text{vesica}, \triangle, \square, \text{pentagon}, \hexagram,
\text{spiral}, \text{none}\}\). For low–moderate \(\mathrm{He}\),
\(P(Y\mid\sigma)\) maximizes at the paired classes above; for broadband/high-\(\mathrm{He}\),
\(P(Y=\text{spiral})\) rises with scale-free field statistics.

**Design (within-subjects)**

- **N=60** naïve participants (balanced; stratify music/geometry exposure).
- **Venues:** 3 materially distinct rooms + anechoic headphone control.
- **Conditions per room:** (1) Mapped (band-limited bursts at measured modal
  peaks) (2) Shuffled (off-peak bins) (3) Broadband high-\(\mathrm{He}\) blocks for
  spiral regime. Separate “authored ritual” sequence tested but analyzed as
  ritual, not law.
- **Loudness:** ISO-226 equal-loudness per participant; SPL logged ±1.5 dB(A).

**Measures**

- **Perceptual:** (i) Free-draw (tablet) (ii) 8-AFC icons (○, vesica, △, □,
  5-star, ✡, spiral, none) with randomized layout (iii) 7-point felt-quality
  sliders (unity/duality/grid/etc.).
- **Physiology:** ECG (RMSSD, LF/HF, DFA α₁/α₂), respiration belt, postural sway
  (force plate/IMU). Optional EEG: envelope-locked SSVEP, 1/f slope.
- **Field:** room RIRs, SPL, temp/humidity; spiral blocks add airflow proxy
  (optical-flow from rigid-mounted camera or hot-wire anemometers).
- **Anti-priming:** no geometry words or images before free-draw.
- **Blinding:** 3 raters blind to condition; κ≥0.70 with mid-study checkpoint;
  decoy inserts to monitor drift.

**Primary hypotheses / fail gates**

- **H1 (Anchors 1–4):** Mapped > Shuffled by ≥12 percentage points in 8-AFC
  accuracy and free-draw clustering agreement (κ≥0.70), across ≥3 rooms; faster
  RTs for anchors.
- **H2 (Portability):** Effects persist across rooms when stimuli are
  room-normalized (no fixed Hz).
- **H3 (Family-specific phys):** Pre-specified signatures differ by class
  (e.g., Stage-2 bi-periodic respiration; Stage-4 sway variance↓) with
  non-significant Room×Family interactions.

**Secondary hypotheses**

- **S1 (Attractors 5/6):** Geometric micro-prime lifts 5/6 accuracy ≥15 pp vs
  baseline; ≥60% retention at 24–48h; HRV DFA α₁ shift Δ≥0.05. Control
  non-geometric prime shows no lift.
- **S2 (Spiral regime):** In ≥2 rooms, broadband blocks jointly show (i)
  sham-subtracted airflow rotation index↑ (ii) physiology scale-free shift
  (HRV DFA α₁↑ toward ~1.0 or EEG 1/f slope) (iii) rotational/expansive language↑
  and odd/even parity flip (Brown > Pink > Control).

**Analysis (Bayesian + frequentist)**

Mixed-effects (participant random intercepts; Room fixed/random as appropriate).
Primary endpoints tested at p<.01 and BF₁₀>10 with ROPEs (e.g., ±3 pp for
accuracy). Image-feature clustering (Radon peaks, rotational spectra, corner
density, Hough lines) cross-checked with human labels (ARI/NMI).

**Controls & exclusions (preregistered)**

- **Ethics:** SPL caps; smoke/schlieren safety; asthma screen; immediate abort
  criteria.
- **Calibration:** SPL meter ref, mic compensation, camera mount spec, IMU
  baseline; randomization seeds logged.
- **Stop rules:** rater κ<0.70 (halt/retrain), SPL drift >±1.5 dB, ambient
  airflow>threshold, drowsiness flags, device failure.
- **Trial exclusion:** motion/breath artifacts (pre-spec thresholds); high
  hypothesis awareness (demand characteristics).

**Outcomes**

- Pass anchors ⇒ support for portable Law spine (1–4 + break).
- Attractors fail ⇒ label 5/6 ritual-contingent.
- Spiral fails ⇒ no emergent regime claim; keep as narrative unless future data
  rescue it.
- All code/data/materials released; deviations logged.

## Stimulus-generation pseudocode (room-normalized, prereg-safe)

```
# --- Calibration & room mapping ---
for ROOM in ROOMS:
  set_up_mic_array()
  set_up_speaker_position()
  sweep_signal = log_sine_sweep(f_start=40Hz, f_end=1200Hz, dur=20s, fs=48k)
  play_and_record(sweep_signal)
  RIR = deconvolve(recording, sweep_signal)
  RT60, FR = estimate_room_response(RIR)
  peaks = find_modal_peaks(FR, min_separation=1/12_octave, Qmin=desired)

  # classify peaks into symmetry families via spatial sampling
  for pk in peaks:
    burst = band_limited_burst(center=pk.freq, bw=1/6_oct, dur=3s, ramp=50ms)
    field = scan_room_microphones(play(burst))
    nodal_pattern = spatial_PSD(field)
    family = classify_family(nodal_pattern)  # radial, 2-lobe, 3-lobe, orthogonal, 5/6, none
    save_mode(ROOM, pk, family)

  # select clean exemplars (lowest interference) per family
  exemplars[ROOM] = choose_one_each(family ∈ {radial,2,3,orthogonal}, criteria=min_cross_talk, stable)

# --- Loudness normalization per participant ---
for PARTICIPANT in participants:
  for ROOM in ROOMS:
    for mode in exemplars[ROOM]:
      level = staircase_MCL(mode.signal)  # ISO-226 compensation
      store_level(PARTICIPANT, ROOM, mode, level)

# --- Build condition playlists ---
for ROOM in ROOMS:
  mapped_list = shuffle(exemplars[ROOM])            # one per family in planned order
  shuffled_list = make_off_peak_bins(FR, avoid=exemplars[ROOM].freqs, match_bandwidth)
  broadband_blocks = {pink_noise, brown_noise, silence_sham}

  # Within-day Latin-square; insert washouts, white-noise resets
  PLAYLIST_DAY_A = interleave(mapped_list, broadband_blocks, catch_trials, latin_square=True)
  PLAYLIST_DAY_B = interleave(shuffled_list, broadband_blocks, catch_trials, latin_square=True)

# --- Trial loop (per room/day) ---
for trial in PLAYLIST:
  set_level(PARTICIPANT, ROOM, trial, stored_level)
  if trial.type == 'broadband':
    ensure_HVAC_off_or_within_spec()
    record_baseline_airflow()            # sham
  play(trial.signal, dur=2–4s)
  record_ECG_resp_IMU()
  capture_free_draw(PEN_TABLET, 10–15s)
  then_forced_choice_8AFC()
  felt_quality_sliders()
  white_noise_wash(3–5s)
  log_SNR_SPL_env()
```

## Free-draw feature-extractor spec (human- & machine-scorable)

**Input**

Vectorized drawing (tablet stylus path) or high-res binary image.

**Preprocessing**

- De-skew & normalize to fixed canvas.
- Morphological open/close to reduce noise.
- Extract largest connected component; compute centroid; normalize scale to unit
  radius.

**Feature families (with thresholds for clustering)**

1. **Radial energy & isotropy**

   - Radial intensity profile \(R(\theta)\) over 0–360°.
   - Uniformity index (U): \(1 - \mathrm{var}(R)/\mathrm{max\_var}\). High U ⇒
     “radial/circle.”
   - Edge curvature mean/variance.

2. **Rotational spectra (FFT over angle)**

   - Compute polar transform; FFT on \(\theta\) to get peaks at harmonics
     \(n=2,3,4,5,6\).
   - Peak prominence \(P_n\) at \(n\in\{2,3,4,5,6\}\); normalize by DC.
   - Max \(P_2\) ⇒ 2-lobe (vesica); \(P_3\) ⇒ triangle; \(P_4\) ⇒ square;
     \(P_5/P_6\) ⇒ star/hex.

3. **Line/vertex structure**

   - Hough transform for straight lines; cluster orientations.
   - Orthogonality score (O): presence of two dominant line families
     ~90°±10°.
   - Corner density (C): FAST/Harris corners per area; high C with O ⇒
     “square/grid.”

4. **Dual-centre detection**

   - Kernel density estimate of stroke density; find local maxima.
   - Bimodality index (B): two peaks with lens-like overlap ⇒ “vesica.”

5. **Star-ray centroid metrics**

   - Skeletonize; count rays emerging from centroid.
   - Ray count (K): K≥5 within ±20° spacing ⇒ “5-ray” (pentagonish); K≈6 ⇒
     “hex.”

6. **Spiral/curvilinear flow**

   - Fit logarithmic spiral \(r = a\cdot e^{b\theta}\) to dominant path (RANSAC).
   - Spiral fit R² and signed curvature persistence (consistent turn direction).
   - Winding number (W): net turns around centroid. High R² + |W|≥0.75 ⇒
     “spiral.”

7. **Global descriptors**

   - Hu moments (scale/rotation invariant), Zernike moments up to order 8.
   - Fractal dimension (box-counting) as auxiliary for spiral/scale-free
     complexity.

**Classification logic (transparent, prereg-able)**

- If U > τ_radial and all \(P_n < \text{small}\_\tau\) ⇒ Circle.
- Else if B > τ_bi and \(P_2\) is dominant ⇒ Vesica.
- Else if \(P_3\) dominant and triangle side straightness > τ ⇒ Triangle.
- Else if O > τ_ortho and \(P_4\) high and corner density C > τ ⇒ Square.
- Else if \(P_5\) dominant and K≥5 ⇒ Pentagonish (5-ray).
- Else if \(P_6\) dominant and K≈6 ⇒ Hexagram/6-cluster.
- Else if Spiral R² > τ_spiral and |W|≥0.75 ⇒ Spiral.
- Else None/Other.

**Human-machine agreement**

- Three blinded raters label drawings using a short rubric mirroring the logic
  above.
- Compute Fleiss’ κ (target ≥0.70) and compare to machine clusters via ARI/NMI.
- Insert 2–3 decoy drawings per batch to monitor rater drift; retrain if batch
  κ<0.70.

**Outputs**

- Primary label + confidence; secondary candidate if close tie (Δscore<ε).
- Feature vector saved (for mixed-effects models & RSA with mode families).

**Quality/artefact checks**

- Reject if stroke length < τ_min or > τ_max (scribble).
- IMU motion spikes within draw window → flag trial.
- Breath/SPL/room covariates logged for later regression.
