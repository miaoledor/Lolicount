import * as THREE from 'three'

/** Original procedural glass jar, hinged lid and candy hearts; no remote model assets. */
export const createHeartJarModel = () => {
  const group = new THREE.Group()
  const glass = new THREE.MeshPhysicalMaterial({
    color: '#ffffff', roughness: 0.08, metalness: 0,
    transparent: true, opacity: 0.12, depthWrite: false,
    side: THREE.DoubleSide, envMapIntensity: 0.7,
  })
  const metal = new THREE.MeshStandardMaterial({ color: '#d7dce3', metalness: 0.94, roughness: 0.24 })
  const cork = new THREE.MeshStandardMaterial({ color: '#c9905a', roughness: 0.85 })
  const profile = [
    [0.001, 0.05], [0.85, 0.05], [1.01, 0.1], [1.07, 0.23],
    [1.07, 2.35], [1.02, 2.58], [1.0, 2.77], [0.94, 2.77],
    [0.96, 2.57], [1.0, 2.34], [1.0, 0.25], [0.86, 0.16], [0.001, 0.16],
  ].map(([x, y]) => new THREE.Vector2(x!, y!))
  const jar = new THREE.Mesh(new THREE.LatheGeometry(profile, 64), glass)
  jar.renderOrder = 2
  group.add(jar)

  const ring = (radius: number, tube: number, y: number, material: THREE.Material = metal) => {
    const mesh = new THREE.Mesh(new THREE.TorusGeometry(radius, tube, 10, 64), material)
    mesh.rotation.x = Math.PI / 2
    mesh.position.y = y
    group.add(mesh)
    return mesh
  }
  ring(0.997, 0.027, 2.76)
  ring(1.015, 0.018, 2.56)
  ring(1.015, 0.025, 0.16, glass)

  const wire = (points: number[][], radius = 0.022) => {
    const curve = new THREE.CatmullRomCurve3(points.map(([x, y, z]) => new THREE.Vector3(x!, y!, z!)))
    return new THREE.Mesh(new THREE.TubeGeometry(curve, 32, radius, 8, false), metal)
  }
  for (const side of [-1, 1]) {
    group.add(wire([
      [side * 1.0, 2.61, 0.06], [side * 1.17, 2.6, 0.08],
      [side * 1.2, 2.42, 0.15], [side * 1.13, 2.37, 0.18],
    ]))
  }

  const lid = new THREE.Group()
  lid.position.set(1.0, 2.57, 0)
  lid.rotation.z = 2.15
  const lidDisk = new THREE.Mesh(new THREE.CylinderGeometry(1.035, 1.035, 0.09, 64), metal)
  lidDisk.position.x = -1
  lid.add(lidDisk)
  const lidCork = new THREE.Mesh(new THREE.CylinderGeometry(0.94, 0.94, 0.1, 64), cork)
  lidCork.position.set(-1, -0.08, 0)
  lid.add(lidCork)
  const lidInset = new THREE.Mesh(new THREE.CylinderGeometry(0.91, 0.91, 0.02, 64), glass)
  lidInset.position.set(-1, 0.06, 0)
  lid.add(lidInset)
  group.add(lid)

  const shape = new THREE.Shape()
  shape.moveTo(0, 0.18)
  shape.bezierCurveTo(-0.35, 0.59, -0.79, 0.1, -0.44, -0.22)
  shape.bezierCurveTo(-0.24, -0.4, 0, -0.58, 0, -0.58)
  shape.bezierCurveTo(0, -0.58, 0.24, -0.4, 0.44, -0.22)
  shape.bezierCurveTo(0.79, 0.1, 0.35, 0.59, 0, 0.18)
  const heartGeometry = new THREE.ExtrudeGeometry(shape, {
    depth: 0.22, bevelEnabled: true, bevelSegments: 3,
    steps: 1, bevelSize: 0.09, bevelThickness: 0.09, curveSegments: 16,
  })
  heartGeometry.center()
  heartGeometry.scale(0.5, 0.5, 0.5)
  const heartMaterials = ['#ff315a', '#ff667b', '#ef2049', '#ff8191', '#e92554'].map((color) =>
    new THREE.MeshPhysicalMaterial({ color, roughness: 0.3, clearcoat: 0.45, clearcoatRoughness: 0.3, envMapIntensity: 0.35 }),
  )
  const createHeart = (index: number) => {
    const mesh = new THREE.Mesh(heartGeometry, heartMaterials[index % heartMaterials.length])
    mesh.rotation.set(Math.sin(index * 1.7) * 0.6, Math.sin(index * 2.3) * 0.9, Math.cos(index * 1.9) * 0.8)
    group.add(mesh)
    return mesh
  }

  const dispose = () => {
    const geometries = new Set<THREE.BufferGeometry>([heartGeometry])
    const materials = new Set<THREE.Material>([glass, metal, cork, ...heartMaterials])
    group.traverse((object) => {
      if (!(object instanceof THREE.Mesh)) return
      geometries.add(object.geometry)
      const list = Array.isArray(object.material) ? object.material : [object.material]
      list.forEach((material) => materials.add(material))
    })
    geometries.forEach((geometry) => geometry.dispose())
    materials.forEach((material) => material.dispose())
  }

  return { group, lid, createHeart, dispose }
}
