// randomNum returns a random integer in the closed interval [min, max].
export const randomNum = (min: number, max: number): number =>
  Math.floor(Math.random() * (max - min + 1) + min)
