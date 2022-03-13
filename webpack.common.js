/**
 * Copyright 2022 Google LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

const HtmlWebpackPlugin = require("html-webpack-plugin");
const FaviconsWebpackPlugin = require("favicons-webpack-plugin");
const CopyPlugin = require("copy-webpack-plugin");
const TsconfigPathsPlugin = require("tsconfig-paths-webpack-plugin");
const webpack = require("webpack");
const packageJson = require("./package.json");
const ForkTsCheckerWebpackPlugin = require("fork-ts-checker-webpack-plugin");

const siteUrl = "https://pi.delivery";

module.exports = {
  plugins: [
    new webpack.DefinePlugin({
      VERSION: JSON.stringify(packageJson.version),
      SITE_URL: JSON.stringify(siteUrl),
    }),
    new HtmlWebpackPlugin({
      template: "_site/index.html",
      scriptLoading: "module",
    }),
    new FaviconsWebpackPlugin({
      logo: "./src/logo.png",
      prefix: "assets/",
      inject: true,
      favicons: {
        appName: packageJson.name,
        appDescription: packageJson.description,
        developerName: "Cloud Developer Relations",
        theme_color: "#0288D1",
        background: "#fff",
        version: packageJson.version,
      },
    }),
    new CopyPlugin({
      patterns: [
        {
          from: "_site",
          globOptions: {
            ignore: ["**/*.html"],
          },
        },
        { from: "public", info: { minimized: true } },
      ],
    }),
    new ForkTsCheckerWebpackPlugin({
      typescript: {
        diagnosticOptions: {
          semantic: true,
          syntactic: true,
        },
      },
    }),
  ],
  entry: {
    index: "./src/index.tsx",
  },
  target: "web",
  module: {
    rules: [
      {
        test: /\.tsx?$/,
        use: ["babel-loader"],
        exclude: /node_modules/,
      },
      {
        test: /\.jsx?$/,
        use: ["babel-loader"],
        exclude: /node_modules/,
      },
      {
        test: /\.s[ac]ss$/,
        use: ["style-loader", "css-loader", "sass-loader"],
      },
      {
        test: /\.css$/,
        use: ["style-loader", "css-loader"],
      },
    ],
  },
  resolve: {
    plugins: [new TsconfigPathsPlugin()],
    extensions: [".tsx", ".ts", ".js"],
  },
  output: {
    clean: true,
    publicPath: "",
  },
};
