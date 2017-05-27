import * as ts from "typescript";

function _go_transpile(source: string): string {
    let result = ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.CommonJS } });
    return result.outputText;
}