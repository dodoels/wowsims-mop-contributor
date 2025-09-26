#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

console.log('🔍 Running pre-release checks...');

// Check if WoWSims executable exists
const executablePath = path.join(__dirname, '..', '..', 'wowsimmop-windows.exe');
if (!fs.existsSync(executablePath)) {
    console.error('❌ FAILED: WoWSims executable not found!');
    console.error(`   Expected: ${executablePath}`);
    console.error('   Please build and copy the WoWSims executable first:');
    console.error('   Run: make wowsimmop-windows.exe in root folder');
    process.exit(1);
}

// Check if TypeScript compiled successfully (with retry for CI)
const mainJsPath = path.join(__dirname, '..', 'dist', 'main.js');
if (!fs.existsSync(mainJsPath)) {
    console.log('⏳ TypeScript files not found, attempting compilation...');
    try {
        const { execSync } = require('child_process');
        execSync('tsc --project . --outDir dist --rootDir src', { cwd: path.join(__dirname, '..'), stdio: 'inherit' });

        // Check again after compilation
        if (!fs.existsSync(mainJsPath)) {
            throw new Error('Compilation completed but files not found');
        }
        console.log('✅ TypeScript compilation successful');
    } catch (error) {
        console.error('❌ FAILED: TypeScript compilation required!');
        console.error('   Please run: npm run compile');
        console.error('   Error:', error.message);
        process.exit(1);
    }
}

// Check if preload script exists
const preloadJsPath = path.join(__dirname, '..', 'dist', 'preload.js');
if (!fs.existsSync(preloadJsPath)) {
    console.error('❌ FAILED: Preload script missing!');
    console.error('   Please run: npm run compile');
    process.exit(1);
}

// Check if assets exist
const iconPath = path.join(__dirname, '..', 'assets', 'WoW-Simulator-Icon.png');
if (!fs.existsSync(iconPath)) {
    console.error('❌ FAILED: Application icon missing!');
    console.error(`   Expected: ${iconPath}`);
    process.exit(1);
}

// Get executable file size for info
const stats = fs.statSync(executablePath);
const fileSizeInMB = (stats.size / (1024 * 1024)).toFixed(2);

console.log('✅ All pre-release checks passed!');
console.log(`   📁 WoWSims executable: ${fileSizeInMB} MB`);
console.log(`   📄 TypeScript compiled: ${fs.existsSync(mainJsPath) ? 'Yes' : 'No'}`);
console.log(`   🎨 Assets ready: ${fs.existsSync(iconPath) ? 'Yes' : 'No'}`);
console.log('🚀 Ready to build release packages!');
