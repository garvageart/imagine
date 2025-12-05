const tz = Intl.supportedValuesOf('timeZone');
console.log('var timezones = []string{');
const chunk = 5; // items per line
for (let i = 0; i < tz.length; i += chunk) {
    const slice = tz.slice(i, i + chunk);
    console.log('\t"' + slice.join('", "') + '",');
}
console.log('}');
