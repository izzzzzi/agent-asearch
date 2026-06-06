// platform detection helper
module.exports = function () {
  return { platform: process.platform, arch: process.arch };
};
