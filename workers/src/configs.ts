// Bot path'leri - tek regex ile hızlı kontrol
const BOT_PATH_REGEX = /\.(env|git|htaccess)|wp-|admin|php|sql|config|backup/i;

// Sadece image/asset uzantılarına izin ver
const ASSET_EXTENSION_REGEX =
  /\.(jpg|jpeg|png|gif|webp|avif|svg|pdf|mp4|webm|mp3|css|js|woff|woff2|ttf|eot|ico)$/i;

const PRESETS = {
  blur: {
    directory: "/blur",
    createUrl: "/cdn/create-blur-image",
  },
};

export { BOT_PATH_REGEX, ASSET_EXTENSION_REGEX, PRESETS };
