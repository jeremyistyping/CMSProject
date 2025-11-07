'use client';

import React from 'react';
import ProtectedModule from '@/components/common/ProtectedModule';
import ProductCatalog from '@/components/products/ProductCatalog';

export default function ProductsPage() {
  return (
    <ProtectedModule module="products">
      <ProductCatalog />
    </ProtectedModule>
  );
}
