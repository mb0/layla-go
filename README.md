layla
=====

layla is a layout and template language based using the xelf project.

It is primarily used as an exploration into the feasibility of the xelf for templates, but should
also provides a simple layout templates for thermal label printers and html previews.

Layout definitions using the xelf as declaration format are already templates with expressions.
Reusing the std lib and some custom specs we can build the layout node tree. The nodes proxy to
custom go structs, making it easy to work with even without using xelf.

Layla supports these layout elements:
      text, block, rect, ellipse, qrcode, barcode elements
      markup with for simple styled text blocks
      stage, group, vbox, hbox and table layouts
      page with extra, cover, header and footer elements for paged documents

There will someday be render packages for:
      tsc   Taiwan Semiconductor (TSC) label printer, specifically for the DA-200 printer
      html  preview in HTML with barcode rendering using boombuler/barcode
      pdf   renderer using jung-kurt/gofpdf

License
-------

Copyright (c) Martin Schnabel. All rights reserved.
Use of the source code is governed by a BSD-style license that can found in the LICENSE file.

This project uses BSD licensed Go fonts for testing (see testdata/README for more info) with
Copyright (c) 2016 Bigelow & Holmes Inc. All rights reserved.
