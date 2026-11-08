export const MAX_HEARTS = 48
export const HEART_RADIUS = 0.27
export const JAR_RADIUS = 0.99
export const JAR_FLOOR = 0.16

export interface HeartBody {
  x: number
  y: number
  z: number
  vx: number
  vy: number
  vz: number
}

export const createHeartBody = (index: number, falling = false): HeartBody => {
  const angle = index * 2.399963
  const radius = falling ? 0.18 : 0.53
  return {
    x: Math.cos(angle) * radius,
    y: falling ? 3.65 : 0.44 + Math.floor(index / 5) * 0.48,
    z: Math.sin(angle) * radius,
    vx: falling ? Math.cos(angle) * 0.3 : 0,
    vy: 0,
    vz: falling ? Math.sin(angle) * 0.3 : 0,
  }
}

/** Sphere contacts approximate the soft hearts without a second physics runtime. */
export const stepHearts = (bodies: HeartBody[], delta: number): void => {
  const dt = Math.min(Math.max(delta, 0), 1 / 30)
  if (dt === 0) return
  for (const body of bodies) {
    body.vy -= 7.5 * dt
    body.x += body.vx * dt
    body.y += body.vy * dt
    body.z += body.vz * dt
  }

  for (let pass = 0; pass < 4; pass++) {
    for (let i = 0; i < bodies.length; i++) {
      const a = bodies[i]!
      for (let j = i + 1; j < bodies.length; j++) {
        const b = bodies[j]!
        const dx = b.x - a.x
        const dy = b.y - a.y
        const dz = b.z - a.z
        const distance = Math.hypot(dx, dy, dz)
        if (distance >= HEART_RADIUS * 2) continue
        const nx = distance > 0.0001 ? dx / distance : 1
        const ny = distance > 0.0001 ? dy / distance : 0
        const nz = distance > 0.0001 ? dz / distance : 0
        const correction = (HEART_RADIUS * 2 - distance) / 2
        a.x -= nx * correction
        a.y -= ny * correction
        a.z -= nz * correction
        b.x += nx * correction
        b.y += ny * correction
        b.z += nz * correction
        const approach = (b.vx - a.vx) * nx + (b.vy - a.vy) * ny + (b.vz - a.vz) * nz
        if (approach < 0) {
          const impulse = -approach * 0.53
          a.vx -= nx * impulse
          a.vy -= ny * impulse
          a.vz -= nz * impulse
          b.vx += nx * impulse
          b.vy += ny * impulse
          b.vz += nz * impulse
        }
      }
    }

    for (const body of bodies) {
      const radius = Math.hypot(body.x, body.z)
      const limit = JAR_RADIUS - HEART_RADIUS
      if (radius > limit) {
        const nx = body.x / radius
        const nz = body.z / radius
        body.x = nx * limit
        body.z = nz * limit
        const outward = body.vx * nx + body.vz * nz
        if (outward > 0) {
          body.vx -= outward * 1.25 * nx
          body.vz -= outward * 1.25 * nz
        }
      }
      if (body.y < JAR_FLOOR + HEART_RADIUS) {
        body.y = JAR_FLOOR + HEART_RADIUS
        body.vy = Math.abs(body.vy) * 0.16
        body.vx *= 0.94
        body.vz *= 0.94
      }
    }
  }
}
