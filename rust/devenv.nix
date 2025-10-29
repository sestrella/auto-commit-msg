{ config, ... }:

{
  env.GEMINI_API_KEY = config.secretspec.secrets.GEMINI_API_KEY or "";

  languages.rust.enable = true;
}
