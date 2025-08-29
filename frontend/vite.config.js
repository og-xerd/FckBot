import { defineConfig } from 'vite';
import obfuscatorPlugin from "vite-plugin-javascript-obfuscator";

export default defineConfig({
  plugins: [
    obfuscatorPlugin({
      compact: true,
      controlFlowFlattening: true,
      controlFlowFlatteningThreshold: 1,
      deadCodeInjection: true,
      deadCodeInjectionThreshold: 1,
      debugProtection: true,
      debugProtectionInterval: true,
      disableConsoleOutput: true,
      identifierNamesGenerator: 'hexadecimal',
      log: false,
      renameGlobals: true,
      selfDefending: true,
      splitStrings: true,
      splitStringsChunkLength: 5,
      stringArray: true,
      stringArrayEncoding: ['rc4'],
      stringArrayThreshold: 1,
      transformObjectKeys: true,
      unicodeEscapeSequence: true,
    })
  ],
  server: {
    host: '0.0.0.0',
    port: 3000,
  }
});