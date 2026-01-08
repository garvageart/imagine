export enum LabelColours {
    Red = "#eb1717",
    Orange = "#f17a18",
    Yellow = "#f5e615",
    Purple = "#9355f7",
    Pink = "#f755a1",
    Green = "#19dd61",
    Blue = "#106ea5",
    None = "transparent"
}

export const flashModes: Record<number, string> = {
    0x0: "Did not fire",
    0x1: "Fired",
    0x5: "Fired, return not detected",
    0x7: "Fired, return detected",
    0x8: "On, did not fire",
    0x9: "On, fired",
    0xD: "On, return not detected",
    0xF: "On, return detected",
    0x10: "Off, did not fire",
    0x14: "Off, did not fire, return not detected",
    0x18: "Auto, did not fire",
    0x19: "Auto, fired",
    0x1D: "Auto, fired, return not detected",
    0x1F: "Auto, fired, return detected",
    0x20: "No flash function",
    0x41: "Fired, red-eye reduction",
    0x45: "Fired, red-eye reduction, return not detected",
    0x47: "Fired, red-eye reduction, return detected",
    0x49: "On, red-eye reduction",
    0x4D: "On, red-eye reduction, fired",
    0x4F: "On, red-eye reduction, return not detected",
    0x50: "Off, red-eye reduction",
    0x58: "Auto, did not fire, red-eye reduction",
    0x59: "Auto, fired, red-eye reduction",
    0x5D: "Auto, fired, return not detected",
    0x5F: "Auto, fired, return detected"
};

export type ImageLabel = keyof typeof LabelColours;