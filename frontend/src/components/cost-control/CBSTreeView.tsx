import React, { useState } from 'react';
import {
    Box,
    VStack,
    HStack,
    Text,
    IconButton,
    Collapse,
    Badge,
    useColorModeValue,
    Progress,
    Tooltip,
    Button,
    Menu,
    MenuButton,
    MenuList,
    MenuItem,
} from '@chakra-ui/react';
import {
    FiChevronRight,
    FiChevronDown,
    FiMoreVertical,
    FiPlus,
    FiEdit2,
    FiTrash2,
    FiDollarSign
} from 'react-icons/fi';
import { CBSNode } from '../../types/cbs';

interface CBSTreeViewProps {
    nodes: CBSNode[];
    onAddNode: (parentId?: number) => void;
    onEditNode: (node: CBSNode) => void;
    onDeleteNode: (node: CBSNode) => void;
    level?: number;
}

const CBSTreeNode: React.FC<{
    node: CBSNode;
    onAddNode: (parentId?: number) => void;
    onEditNode: (node: CBSNode) => void;
    onDeleteNode: (node: CBSNode) => void;
    level: number;
}> = ({ node, onAddNode, onEditNode, onDeleteNode, level }) => {
    const [isOpen, setIsOpen] = useState(true);
    const hasChildren = node.children && node.children.length > 0;

    const bgHover = useColorModeValue('gray.50', 'gray.700');
    const borderColor = useColorModeValue('gray.200', 'gray.600');

    // Calculate budget usage percentage
    const usagePercent = node.budget_amount > 0
        ? (node.actual_cost / node.budget_amount) * 100
        : 0;

    // Determine color based on usage
    let progressColor = 'green';
    if (usagePercent > 90) progressColor = 'red';
    else if (usagePercent > 75) progressColor = 'yellow';

    const formatCurrency = (amount: number) => {
        return new Intl.NumberFormat('id-ID', {
            style: 'currency',
            currency: 'IDR',
            minimumFractionDigits: 0,
            maximumFractionDigits: 0,
        }).format(amount);
    };

    return (
        <Box>
            <HStack
                spacing={2}
                p={2}
                pl={`${level * 20 + 8}px`}
                borderBottomWidth="1px"
                borderColor={borderColor}
                _hover={{ bg: bgHover }}
                role="group"
            >
                <IconButton
                    aria-label="Toggle expand"
                    icon={isOpen ? <FiChevronDown /> : <FiChevronRight />}
                    size="xs"
                    variant="ghost"
                    visibility={hasChildren ? 'visible' : 'hidden'}
                    onClick={() => setIsOpen(!isOpen)}
                />

                <VStack align="start" spacing={0} flex={1}>
                    <HStack>
                        <Badge colorScheme="blue" fontSize="xs">{node.code}</Badge>
                        <Text fontWeight="medium" fontSize="sm">{node.name}</Text>
                    </HStack>
                    {node.description && (
                        <Text fontSize="xs" color="gray.500" noOfLines={1}>{node.description}</Text>
                    )}
                </VStack>

                {/* Budget & Cost Info */}
                <VStack align="end" spacing={0} w="200px" display={{ base: 'none', md: 'flex' }}>
                    <Text fontSize="xs" color="gray.500">Budget: {formatCurrency(node.budget_amount)}</Text>
                    <Text fontSize="xs" fontWeight="bold" color={usagePercent > 100 ? 'red.500' : 'green.500'}>
                        Actual: {formatCurrency(node.actual_cost)}
                    </Text>
                </VStack>

                {/* Progress Bar */}
                <Box w="100px" display={{ base: 'none', md: 'block' }}>
                    <Tooltip label={`${usagePercent.toFixed(1)}% Used`}>
                        <Progress
                            value={Math.min(usagePercent, 100)}
                            size="sm"
                            colorScheme={progressColor}
                            borderRadius="full"
                        />
                    </Tooltip>
                </Box>

                {/* Actions */}
                <Menu>
                    <MenuButton
                        as={IconButton}
                        icon={<FiMoreVertical />}
                        variant="ghost"
                        size="sm"
                        opacity={0}
                        _groupHover={{ opacity: 1 }}
                    />
                    <MenuList>
                        <MenuItem icon={<FiPlus />} onClick={() => onAddNode(node.id)}>
                            Add Sub-item
                        </MenuItem>
                        <MenuItem icon={<FiEdit2 />} onClick={() => onEditNode(node)}>
                            Edit
                        </MenuItem>
                        <MenuItem icon={<FiTrash2 />} onClick={() => onDeleteNode(node)} color="red.500">
                            Delete
                        </MenuItem>
                    </MenuList>
                </Menu>
            </HStack>

            <Collapse in={isOpen} animateOpacity>
                {hasChildren && node.children?.map((child) => (
                    <CBSTreeNode
                        key={child.id}
                        node={child}
                        onAddNode={onAddNode}
                        onEditNode={onEditNode}
                        onDeleteNode={onDeleteNode}
                        level={level + 1}
                    />
                ))}
            </Collapse>
        </Box>
    );
};

const CBSTreeView: React.FC<CBSTreeViewProps> = ({ nodes, onAddNode, onEditNode, onDeleteNode }) => {
    const borderColor = useColorModeValue('gray.200', 'gray.600');

    if (!nodes || nodes.length === 0) {
        return (
            <Box p={8} textAlign="center" borderWidth="1px" borderRadius="lg" borderStyle="dashed">
                <Text color="gray.500">No Cost Breakdown Structure defined yet.</Text>
                <Text color="gray.400" fontSize="sm" mt={2}>Use the "Create Root Node" button above to get started.</Text>
            </Box>
        );
    }

    return (
        <Box borderWidth="1px" borderRadius="lg" borderColor={borderColor} overflow="hidden">
            <HStack bg={useColorModeValue('gray.50', 'gray.700')} p={3} borderBottomWidth="1px" borderColor={borderColor}>
                <Text fontWeight="bold" fontSize="sm" flex={1} pl={8}>Cost Code / Description</Text>
                <Text fontWeight="bold" fontSize="sm" w="200px" textAlign="right" display={{ base: 'none', md: 'block' }}>Financials</Text>
                <Text fontWeight="bold" fontSize="sm" w="100px" textAlign="center" display={{ base: 'none', md: 'block' }}>Usage</Text>
                <Box w="40px" /> {/* Spacer for actions */}
            </HStack>

            {nodes.map((node) => (
                <CBSTreeNode
                    key={node.id}
                    node={node}
                    onAddNode={onAddNode}
                    onEditNode={onEditNode}
                    onDeleteNode={onDeleteNode}
                    level={0}
                />
            ))}
        </Box>
    );
};

export default CBSTreeView;
