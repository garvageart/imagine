import { DateTime } from "luxon";

export enum LogLevel {
    DEBUG = 'DEBUG',
    INFO = 'INFO',
    LOG = 'LOG',
    WARN = 'WARN',
    ERROR = 'ERROR',
}

/**
 * Used in dev mode to nicely log to the console
 */
export const logger = (
    level: LogLevel,
    message: string | Error,
    ...args: any
) => {
    const colorize = withColor(
        {
            [LogLevel.DEBUG]: colors.magenta,
            [LogLevel.INFO]: colors.blue,
            [LogLevel.LOG]: colors.green,
            [LogLevel.WARN]: colors.yellow,
            [LogLevel.ERROR]: colors.red,
        }[level],
    );

    let msg = `${colorize(DateTime.now().toFormat("dd-MM-yyyy HH:mm:ss"))}`;
    if (message) {
        msg += ` ${message.toString()}`;
    }

    if (message instanceof Error) {
        // for stacktrace
        console.error(message);
    } else {
        console.log(msg, ...args);
    }
};

/**
 * Color handling copied from Kleur
 *
 * @see https://github.com/lukeed/kleur/blob/fa3454483899ddab550d08c18c028e6db1aab0e5/colors.mjs#L13
 */
const colors: {
    [key: string]: [number, number];
} = {
    default: [0, 0],
    blue: [36, 39],
    green: [32, 39],
    magenta: [35, 39],
    red: [31, 39],
    yellow: [33, 39],
};

function withColor([x, y]: [number, number]) {
    const rgx = new RegExp(`\\x1b\\[${y}m`, 'g');
    const open = `\x1b[${x}m`,
        close = `\x1b[${y}m`;

    return function (txt: string) {
        if (txt == null) return txt;

        return (
            open +
            (~('' + txt).indexOf(close) ? txt.replace(rgx, close + open) : txt) +
            close
        );
    };
}