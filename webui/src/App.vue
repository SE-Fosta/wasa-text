<script setup>
// Non serve logica qui, il router gestisce tutto.
</script>

<template>
  <router-view />
</template>

<style>
/* Stili specifici per il contenitore principale #app */
#app {
  width: 100%;
  height: 100vh; /* Forza l'altezza al 100% della finestra */
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
}
</style>
