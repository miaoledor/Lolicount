import * as THREE from 'three'
import { RoomEnvironment } from 'three/addons/environments/RoomEnvironment.js'
import { createHeartJarModel } from './heartJarModel'
import { createHeartBody, MAX_HEARTS, stepHearts, type HeartBody } from './heartPhysics'

/** Owns the WebGL lifetime and renders only while visible and moving. */
export const createHeartJarScene = (host: HTMLElement, reducedMotion: boolean) => {
  const renderer = new THREE.WebGLRenderer({ alpha: true, antialias: true, powerPreference: 'low-power' })
  renderer.setPixelRatio(Math.min(window.devicePixelRatio, 1.75))
  renderer.setClearColor(0x000000, 0)
  renderer.toneMapping = THREE.ACESFilmicToneMapping
  renderer.toneMappingExposure = 0.95
  renderer.domElement.setAttribute('aria-hidden', 'true')
  host.append(renderer.domElement)

  const scene = new THREE.Scene()
  const camera = new THREE.PerspectiveCamera(33, 1, 0.1, 30)
  camera.position.set(0.7, 3.8, 8.6)
  camera.lookAt(0.45, 1.65, 0)
  const pmrem = new THREE.PMREMGenerator(renderer)
  const room = new RoomEnvironment()
  const environment = pmrem.fromScene(room, 0.04)
  scene.environment = environment.texture
  room.dispose()
  pmrem.dispose()
  scene.add(new THREE.HemisphereLight(0xffffff, 0x9c86a5, 1.5))
  const key = new THREE.DirectionalLight(0xfff4ec, 2.5)
  key.position.set(-3, 6, 5)
  scene.add(key)
  const rim = new THREE.DirectionalLight(0xb5c8ff, 1)
  rim.position.set(3, 4, -3)
  scene.add(rim)
  const model = createHeartJarModel()
  scene.add(model.group)
  const bodies: HeartBody[] = []
  const meshes: THREE.Mesh[] = []
  let serial = 0
  let yaw = -0.15
  let targetYaw = yaw
  let frame = 0
  let previousTime = 0
  let activeUntil = 0
  let visible = true
  let disposed = false
  let reduce = reducedMotion

  const sync = () => {
    bodies.forEach((body, index) => meshes[index]!.position.set(body.x, body.y, body.z))
    model.group.rotation.y = yaw
  }
  const settle = () => {
    for (let i = 0; i < 240; i++) stepHearts(bodies, 1 / 60)
  }
  const render = (time: number) => {
    frame = 0
    if (disposed || !visible || document.hidden) return
    const delta = previousTime ? Math.min((time - previousTime) / 1000, 1 / 30) : 1 / 60
    previousTime = time
    if (!reduce) {
      stepHearts(bodies, delta / 2)
      stepHearts(bodies, delta / 2)
    }
    yaw += (targetYaw - yaw) * (reduce ? 1 : 0.12)
    sync()
    renderer.render(scene, camera)
    if (!reduce && time < activeUntil) frame = requestAnimationFrame(render)
  }
  const invalidate = (duration = 4000) => {
    activeUntil = performance.now() + duration
    if (!frame && visible && !document.hidden && !disposed) {
      previousTime = 0
      frame = requestAnimationFrame(render)
    }
  }
  const addHeart = (falling = true) => {
    if (bodies.length >= MAX_HEARTS) {
      bodies.shift()
      model.group.remove(meshes.shift()!)
    }
    bodies.push(createHeartBody(serial, falling))
    meshes.push(model.createHeart(serial++))
    if (reduce) settle()
    invalidate()
  }
  for (let i = 0; i < 20; i++) addHeart(false)
  settle()

  const resize = () => {
    const { width, height } = host.getBoundingClientRect()
    if (!width || !height) return
    renderer.setSize(width, height)
    camera.aspect = width / height
    camera.updateProjectionMatrix()
    invalidate(0)
  }
  const observer = new ResizeObserver(resize)
  observer.observe(host)
  const setVisible = (value: boolean) => {
    visible = value
    if (value) invalidate()
    else {
      cancelAnimationFrame(frame)
      frame = 0
    }
  }
  const visibility = () => setVisible(!document.hidden && inView)
  let inView = true
  const intersection = new IntersectionObserver(([entry]) => {
    inView = !!entry?.isIntersecting
    setVisible(inView && !document.hidden)
  })
  intersection.observe(host)
  document.addEventListener('visibilitychange', visibility)
  resize()

  return {
    addHeart,
    rotate: (amount: number) => { targetYaw += amount; invalidate(900) },
    setReducedMotion: (value: boolean) => {
      reduce = value
      if (reduce) settle()
      invalidate()
    },
    dispose: () => {
      disposed = true
      cancelAnimationFrame(frame)
      observer.disconnect()
      intersection.disconnect()
      document.removeEventListener('visibilitychange', visibility)
      model.dispose()
      environment.dispose()
      renderer.dispose()
      renderer.forceContextLoss()
      renderer.domElement.remove()
    },
  }
}

export type HeartJarScene = ReturnType<typeof createHeartJarScene>
