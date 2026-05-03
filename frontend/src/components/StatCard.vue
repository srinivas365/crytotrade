<template>
  <div class="bg-white rounded-xl border border-gray-200 p-5">
    <p class="text-xs font-semibold text-gray-500 uppercase tracking-wide">{{ label }}</p>
    <p class="text-3xl font-bold mt-1" :class="valueClass">{{ value }}</p>
    <p v-if="sub" class="text-xs text-gray-400 mt-1">{{ sub }}</p>
  </div>
</template>

<script setup>
defineProps({
  label: String,
  value: [String, Number],
  sub: String,
  valueClass: { type: String, default: 'text-gray-900' },
})
</script>
