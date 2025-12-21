import type { MaterialSymbol } from "material-symbols";

export interface ShareSettings {

}

export interface AdvancedSettings {
    
}

export interface Settings {
    share: ShareSettings
    advanced: AdvancedSettings
}

export type DropdownOption = {
    title: string;
    icon?: MaterialSymbol;
    disabled?: boolean;
};