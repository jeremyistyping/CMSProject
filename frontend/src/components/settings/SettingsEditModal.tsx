'use client';

import React, { useState, useEffect } from 'react';
import {
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalFooter,
  ModalBody,
  ModalCloseButton,
  Button,
  FormControl,
  FormLabel,
  Input,
  VStack,
  HStack,
  Tabs,
  TabList,
  TabPanels,
  Tab,
  TabPanel,
  Select,
  Switch,
  useToast,
  FormErrorMessage,
  Textarea,
  NumberInput,
  NumberInputField,
  NumberInputStepper,
  NumberIncrementStepper,
  NumberDecrementStepper,
  Tooltip,
  Icon,
} from '@chakra-ui/react';
import { FiInfo } from 'react-icons/fi';
import { useTranslation } from '@/hooks/useTranslation';
import api from '@/services/api';

interface SystemSettings {
  id?: number;
  company_name: string;
  company_address: string;
  company_phone: string;
  company_email: string;
  company_website?: string;
  tax_number?: string;
  currency: string;
  date_format: string;
  fiscal_year_start: string;
  default_tax_rate: number;
  language: string;
  timezone: string;
  thousand_separator: string;
  decimal_separator: string;
  decimal_places: number;
  invoice_prefix: string;
  invoice_next_number?: number; // optional: field removed from API
  quote_prefix: string;
  quote_next_number?: number; // optional: field removed from API
  purchase_prefix: string;
  purchase_next_number?: number; // optional: field removed from API
  payment_receivable_prefix?: string;
  payment_payable_prefix?: string;
}

interface SettingsEditModalProps {
  isOpen: boolean;
  onClose: () => void;
  settings: SystemSettings | null;
  onUpdate: (settings: SystemSettings) => void;
}

const SettingsEditModal: React.FC<SettingsEditModalProps> = ({
  isOpen,
  onClose,
  settings,
  onUpdate,
}) => {
  const { t } = useTranslation();
  const toast = useToast();
  const [formData, setFormData] = useState<SystemSettings>({
    company_name: '',
    company_address: '',
    company_phone: '',
    company_email: '',
    company_website: '',
    tax_number: '',
    currency: 'IDR',
    date_format: 'DD/MM/YYYY',
    fiscal_year_start: 'January 1',
    default_tax_rate: 11,
    language: 'id',
    timezone: 'Asia/Jakarta',
    thousand_separator: '.',
    decimal_separator: ',',
    decimal_places: 2,
    invoice_prefix: 'INV',
    invoice_next_number: 1,
    quote_prefix: 'QT',
    quote_next_number: 1,
    purchase_prefix: 'PO',
    purchase_next_number: 1,
    payment_receivable_prefix: 'RCV',
    payment_payable_prefix: 'PAY',
  });
  const [isLoading, setIsLoading] = useState(false);
  const [errors, setErrors] = useState<Partial<Record<keyof SystemSettings, string>>>({});

  useEffect(() => {
    if (settings) {
      setFormData(settings);
    }
  }, [settings]);

  const validateForm = (): boolean => {
    const newErrors: Partial<Record<keyof SystemSettings, string>> = {};

    if (!formData.company_name) {
      newErrors.company_name = t('settings.validation.companyNameRequired');
    }

    if (!formData.company_email) {
      newErrors.company_email = t('settings.validation.emailRequired');
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(formData.company_email)) {
      newErrors.company_email = t('settings.validation.invalidEmail');
    }

    if (formData.default_tax_rate < 0 || formData.default_tax_rate > 100) {
      newErrors.default_tax_rate = t('settings.validation.invalidTaxRate');
    }

    if (formData.decimal_places < 0 || formData.decimal_places > 4) {
      newErrors.decimal_places = t('settings.validation.invalidDecimalPlaces');
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async () => {
    if (!validateForm()) {
      return;
    }

    setIsLoading(true);
    try {
      const response = await api.put('/api/v1/settings', formData);
      
      if (response.data.success) {
        toast({
          title: t('settings.updateSuccess'),
          status: 'success',
          duration: 3000,
          isClosable: true,
        });
        onUpdate(response.data.data);
        onClose();
      }
    } catch (error: any) {
      toast({
        title: t('settings.updateError'),
        description: error.response?.data?.details || error.message,
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    } finally {
      setIsLoading(false);
    }
  };

  const handleInputChange = (field: keyof SystemSettings, value: any) => {
    setFormData(prev => ({
      ...prev,
      [field]: value,
    }));
    // Clear error for this field
    if (errors[field]) {
      setErrors(prev => {
        const newErrors = { ...prev };
        delete newErrors[field];
        return newErrors;
      });
    }
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} size="xl" scrollBehavior="inside">
      <ModalOverlay />
      <ModalContent maxW="900px">
        <ModalHeader>{t('settings.editSettings')}</ModalHeader>
        <ModalCloseButton />
        <ModalBody>
          <Tabs>
            <TabList>
              <Tab>
                <HStack spacing={2}>
                  <span>{t('settings.companyInfo')}</span>
                  <Tooltip
                    label={t('settings.tooltips.companyInfo') || 'Configure your company information including name, address, contact details, and tax number. This information will appear on invoices and reports.'}
                    placement="top"
                    hasArrow
                  >
                    <Icon as={FiInfo} boxSize={3} color="blue.500" cursor="help" />
                  </Tooltip>
                </HStack>
              </Tab>
              <Tab>
                <HStack spacing={2}>
                  <span>{t('settings.systemConfig')}</span>
                  <Tooltip
                    label={t('settings.tooltips.systemConfig') || 'System-wide configuration including language, timezone, number formats, and fiscal year settings.'}
                    placement="top"
                    hasArrow
                  >
                    <Icon as={FiInfo} boxSize={3} color="green.500" cursor="help" />
                  </Tooltip>
                </HStack>
              </Tab>
              <Tab>
                <HStack spacing={2}>
                  <span>{t('settings.invoiceSettings')}</span>
                  <Tooltip
                    label={t('settings.tooltips.invoiceSettings') || 'Configure numbering systems for invoices, quotes, and purchase orders. Set prefixes and next numbers for automatic generation.'}
                    placement="top"
                    hasArrow
                  >
                    <Icon as={FiInfo} boxSize={3} color="purple.500" cursor="help" />
                  </Tooltip>
                </HStack>
              </Tab>
            </TabList>

            <TabPanels>
              {/* Company Information Tab */}
              <TabPanel>
                <VStack spacing={4}>
                  <FormControl isInvalid={!!errors.company_name} isRequired>
                    <FormLabel>{t('settings.companyName')}</FormLabel>
                    <Input
                      value={formData.company_name}
                      onChange={(e) => handleInputChange('company_name', e.target.value)}
                      placeholder={t('settings.companyNamePlaceholder')}
                    />
                    <FormErrorMessage>{errors.company_name}</FormErrorMessage>
                  </FormControl>

                  <FormControl>
                    <FormLabel>{t('settings.address')}</FormLabel>
                    <Textarea
                      value={formData.company_address}
                      onChange={(e) => handleInputChange('company_address', e.target.value)}
                      placeholder={t('settings.addressPlaceholder')}
                      rows={3}
                    />
                  </FormControl>

                  <HStack width="full" spacing={4}>
                    <FormControl flex={1}>
                      <FormLabel>{t('settings.phone')}</FormLabel>
                      <Input
                        value={formData.company_phone}
                        onChange={(e) => handleInputChange('company_phone', e.target.value)}
                        placeholder={t('settings.phonePlaceholder')}
                      />
                    </FormControl>

                    <FormControl flex={1} isInvalid={!!errors.company_email} isRequired>
                      <FormLabel>{t('settings.email')}</FormLabel>
                      <Input
                        type="email"
                        value={formData.company_email}
                        onChange={(e) => handleInputChange('company_email', e.target.value)}
                        placeholder={t('settings.emailPlaceholder')}
                      />
                      <FormErrorMessage>{errors.company_email}</FormErrorMessage>
                    </FormControl>
                  </HStack>

                  <HStack width="full" spacing={4}>
                    <FormControl flex={1}>
                      <FormLabel>{t('settings.website')}</FormLabel>
                      <Input
                        value={formData.company_website || ''}
                        onChange={(e) => handleInputChange('company_website', e.target.value)}
                        placeholder={t('settings.websitePlaceholder')}
                      />
                    </FormControl>

                    <FormControl flex={1}>
                      <FormLabel>{t('settings.taxNumber')}</FormLabel>
                      <Input
                        value={formData.tax_number || ''}
                        onChange={(e) => handleInputChange('tax_number', e.target.value)}
                        placeholder={t('settings.taxNumberPlaceholder')}
                      />
                    </FormControl>
                  </HStack>
                </VStack>
              </TabPanel>

              {/* System Configuration Tab */}
              <TabPanel>
                <VStack spacing={4}>
                  <HStack width="full" spacing={4}>
                    <FormControl flex={1}>
                      <FormLabel>{t('settings.language')}</FormLabel>
                      <Select
                        value={formData.language}
                        onChange={(e) => handleInputChange('language', e.target.value)}
                      >
                        <option value="id">{t('settings.indonesian')}</option>
                        <option value="en">{t('settings.english')}</option>
                      </Select>
                    </FormControl>

                    <FormControl flex={1}>
                    </FormControl>
                  </HStack>

                  <HStack width="full" spacing={4}>
                    <FormControl flex={1}>
                      <FormLabel>{t('settings.dateFormat')}</FormLabel>
                      <Select
                        value={formData.date_format}
                        onChange={(e) => handleInputChange('date_format', e.target.value)}
                      >
                        <option value="DD/MM/YYYY">DD/MM/YYYY</option>
                        <option value="MM/DD/YYYY">MM/DD/YYYY</option>
                        <option value="YYYY-MM-DD">YYYY-MM-DD</option>
                      </Select>
                    </FormControl>

                    <FormControl flex={1}>
                      <FormLabel>
                        <HStack spacing={1}>
                          <span>{t('settings.fiscalYearStart')}</span>
                          <Tooltip
                            label="Start date of your company's fiscal year. Affects financial reporting periods and year-end calculations."
                            placement="top"
                            hasArrow
                          >
                            <Icon as={FiInfo} boxSize={3} color="blue.400" cursor="help" />
                          </Tooltip>
                        </HStack>
                      </FormLabel>
                      <Select
                        value={formData.fiscal_year_start}
                        onChange={(e) => handleInputChange('fiscal_year_start', e.target.value)}
                      >
                        <option value="January 1">January 1</option>
                        <option value="April 1">April 1</option>
                        <option value="July 1">July 1</option>
                        <option value="October 1">October 1</option>
                      </Select>
                    </FormControl>
                  </HStack>

                  <HStack width="full" spacing={4}>
                  <FormControl flex={1} isInvalid={!!errors.default_tax_rate}>
                      <FormLabel>
                        <HStack spacing={1}>
                          <span>{t('settings.defaultTaxRate')} (%)</span>
                          <Tooltip
                            label="Default tax rate applied to transactions. In Indonesia, standard VAT is 11%. Range: 0-100%"
                            placement="top"
                            hasArrow
                          >
                            <Icon as={FiInfo} boxSize={3} color="blue.400" cursor="help" />
                          </Tooltip>
                        </HStack>
                      </FormLabel>
                      <NumberInput
                        value={formData.default_tax_rate}
                        onChange={(_, value) => handleInputChange('default_tax_rate', value)}
                        min={0}
                        max={100}
                        precision={2}
                      >
                        <NumberInputField />
                        <NumberInputStepper>
                          <NumberIncrementStepper />
                          <NumberDecrementStepper />
                        </NumberInputStepper>
                      </NumberInput>
                      <FormErrorMessage>{errors.default_tax_rate}</FormErrorMessage>
                    </FormControl>

                    <FormControl flex={1}>
                      <FormLabel>{t('settings.timezone')}</FormLabel>
                      <Select
                        value={formData.timezone}
                        onChange={(e) => handleInputChange('timezone', e.target.value)}
                      >
                        <option value="Asia/Jakarta">Asia/Jakarta (WIB)</option>
                        <option value="Asia/Makassar">Asia/Makassar (WITA)</option>
                        <option value="Asia/Jayapura">Asia/Jayapura (WIT)</option>
                      </Select>
                    </FormControl>
                  </HStack>

                  <HStack width="full" spacing={4}>
                    <FormControl flex={1}>
                      <FormLabel>{t('settings.thousandSeparator')}</FormLabel>
                      <Select
                        value={formData.thousand_separator}
                        onChange={(e) => handleInputChange('thousand_separator', e.target.value)}
                      >
                        <option value=".">. (Period)</option>
                        <option value=",">, (Comma)</option>
                        <option value=" ">  (Space)</option>
                      </Select>
                    </FormControl>

                    <FormControl flex={1}>
                      <FormLabel>{t('settings.decimalSeparator')}</FormLabel>
                      <Select
                        value={formData.decimal_separator}
                        onChange={(e) => handleInputChange('decimal_separator', e.target.value)}
                      >
                        <option value=",">, (Comma)</option>
                        <option value=".">. (Period)</option>
                      </Select>
                    </FormControl>

                    <FormControl flex={1} isInvalid={!!errors.decimal_places}>
                      <FormLabel>{t('settings.decimalPlaces')}</FormLabel>
                      <NumberInput
                        value={formData.decimal_places}
                        onChange={(_, value) => handleInputChange('decimal_places', value)}
                        min={0}
                        max={4}
                      >
                        <NumberInputField />
                        <NumberInputStepper>
                          <NumberIncrementStepper />
                          <NumberDecrementStepper />
                        </NumberInputStepper>
                      </NumberInput>
                      <FormErrorMessage>{errors.decimal_places}</FormErrorMessage>
                    </FormControl>
                  </HStack>
                </VStack>
              </TabPanel>

              {/* Invoice Settings Tab */}
              <TabPanel>
                <VStack spacing={4}>
                  <HStack width="full" spacing={4}>
                    <FormControl flex={1}>
                      <FormLabel>{t('settings.invoicePrefix')}</FormLabel>
                      <Input
                        value={formData.invoice_prefix}
                        onChange={(e) => handleInputChange('invoice_prefix', e.target.value)}
                        placeholder="INV"
                      />
                    </FormControl>

                  </HStack>


                  <HStack width="full" spacing={4}>
                    <FormControl flex={1}>
                      <FormLabel>{t('settings.purchasePrefix')}</FormLabel>
                      <Input
                        value={formData.purchase_prefix}
                        onChange={(e) => handleInputChange('purchase_prefix', e.target.value)}
                        placeholder="PO"
                      />
                    </FormControl>

                  </HStack>

                  <HStack width="full" spacing={4}>
                    <FormControl flex={1}>
                      <FormLabel>{t('settings.paymentReceivablePrefix')}</FormLabel>
                      <Input
                        value={formData.payment_receivable_prefix || ''}
                        onChange={(e) => handleInputChange('payment_receivable_prefix', e.target.value)}
                        placeholder="RCV"
                      />
                    </FormControl>

                    <FormControl flex={1}>
                      <FormLabel>{t('settings.paymentPayablePrefix')}</FormLabel>
                      <Input
                        value={formData.payment_payable_prefix || ''}
                        onChange={(e) => handleInputChange('payment_payable_prefix', e.target.value)}
                        placeholder="PAY"
                      />
                    </FormControl>
                  </HStack>
                </VStack>
              </TabPanel>
            </TabPanels>
          </Tabs>
        </ModalBody>

        <ModalFooter>
          <Button variant="ghost" mr={3} onClick={onClose} isDisabled={isLoading}>
            {t('common.cancel')}
          </Button>
          <Button colorScheme="blue" onClick={handleSubmit} isLoading={isLoading}>
            {t('common.save')}
          </Button>
        </ModalFooter>
      </ModalContent>
    </Modal>
  );
};

export default SettingsEditModal;
