import { defineConfig } from '@kubb/core';
import { pluginOas } from '@kubb/plugin-oas';
import { pluginTs } from '@kubb/plugin-ts';
import { pluginReactQuery } from '@kubb/plugin-react-query';
import { pluginClient } from '@kubb/plugin-client';

export default defineConfig({
  root: '.',
  input: {
    path: '../backend/openapi/bundled.yaml',
  },
  output: {
    path: './src/gen',
    clean: true,
  },
  plugins: [
    pluginOas({}),
    pluginTs({}),
    pluginClient({
      importPath: '../../client.ts',
    }),
    pluginReactQuery({
      output: {
        path: './hooks',
      },
    }),
  ],
});
