import { useMouse } from '@vueuse/core'

export default {
  setup() {
    const { x, y, sourceType } = useMouse()
    return { x, y, sourceType };
  },
  template: /*html*/`
    <div>
      <p>x: {{ x }}</p>
      <p>y: {{ y }}</p>
      <p>sourceType: {{ sourceType }}</p>
    </div>
  `,
};