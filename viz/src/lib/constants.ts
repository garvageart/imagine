import { dev } from "$app/environment";

const IS_BROWSER_ENV = {
    production: dev === false || (location.port === '') || (location.hostname !== 'localhost'),
    development: dev === true || (location.port !== '') || (location.hostname === 'localhost'),
};

export const IS_MOBILE = /iPhone|iPad|iPod|Android/i.test(navigator.userAgent) || screen.orientation.type === 'portrait-primary';
export const CLIENT_IS_PRODUCTION = IS_BROWSER_ENV?.production;
export const BROWSER_BASE_URL = window.location.hostname;

export class ServerURLConfig {
    host: string;
    port: number;
    url: string;
    subdomain: string;

    constructor(
        subdomain: string,
        port: number
    ) {
        this.port = port;
        this.subdomain = subdomain;

        if (CLIENT_IS_PRODUCTION) {
            this.host = `https://${subdomain}.imagine.${BROWSER_BASE_URL}`;
        } else {
            this.host = `http://localhost:${port}`;
        }

        this.url = this.host;
    }
}

export const MEDIA_SERVER = new ServerURLConfig("media", 7770);
export const AUTH_SERVER = new ServerURLConfig("auth", 7771);
export const UI_SERVER = new ServerURLConfig("viz", 7777);

export const DEFAULT_THEME = "viz-theme";
export const CAN_DEBUG = IS_BROWSER_ENV.development && localStorage.getItem("viz:debugMode") === "true";
