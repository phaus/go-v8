"use strict";

function _go_transpile(source) {
    var result = ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.CommonJS } });
    return result.outputText;
}
