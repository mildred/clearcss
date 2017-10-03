clearcss
========

clearcss is a CSS preprocessor that allows you reuse while keeping a clear
separation between markup and styling. CSS frameworks like bootstrap will define
HTML classes that are style-oriented and not semantic. clearcss will let you
keep your semantic markup.

It works by defining two directives that will be replaced:

@require
--------

- Usage: `@require "<relative path to stylesheet>";`
- Context: anywhere

Declares a dependency to another stylesheet. The stylesheet is analyzed and
ready to be reused. The stylesheet is not imported and will not appear in the
final result.

@extend
-------

- Usage: `@extend <CSS selector>, <CSS selector>, ...;`
- Context: inside another selector

Extends the current selector with the selector specified. The CSS properties
corresponding to this selector that are in the required stylesheets are put in
place of the `@extend` directive. Helpful comments are added.

Build
=====

    go build .

Install
=======

    go install .

Usage
=====

    clearcss STYLESHEET.css >OUTPUT.css

Will read `STYLESHEET.css` and put the processed stylesheet in `OUTPUT.css`
(done by shell redirection)

Example
=======

Take this example stylesheet:

    @require "bootstrap/dist/css/bootstrap.css";

    html {
      @extend html;
    }

Executed with `clearcss example.in.css`, this results in:

    html {
      -webkit-box-sizing: border-box;
      box-sizing: border-box;
      font-size: 16px;
      -ms-overflow-style: scrollbar;
      -webkit-tap-highlight-color: transparent;
    }

